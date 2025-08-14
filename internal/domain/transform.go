package domain

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/bufbuild/protocompile"
	"github.com/cockroachdb/errors"
	"github.com/rs/zerolog"
	"github.com/samber/lo"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	"github.com/Dmytro-Hladkykh/gripmock/internal/pbs"
)

const (
	ProtoExt       = ".proto"
	ProtobufSetExt = ".pb"
	ProtoSetExt    = ".protoset"

	fileTypeProto      = "proto"
	fileTypeDescriptor = "descriptor"
)

var errUnsupportedFileType = errors.New("unsupported file type")

type Configure struct {
	imports     []string
	protos      []string
	descriptors []string
}

func (c *Configure) Imports() []string     { return c.imports }
func (c *Configure) Protos() []string      { return c.protos }
func (c *Configure) Descriptors() []string { return c.descriptors }

func createDescriptorSet(ctx context.Context, configure *Configure) (*descriptorpb.FileDescriptorSet, error) {
	failbackResolver, err := pbs.NewResolver()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create fallback resolver")
	}

	compiler := protocompile.Compiler{
		Resolver: protocompile.CompositeResolver{
			&protocompile.SourceResolver{
				ImportPaths: configure.Imports(),
			},
			failbackResolver,
		},
	}

	files, err := compiler.Compile(ctx, configure.Protos()...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile descriptors")
	}

	fds := &descriptorpb.FileDescriptorSet{
		File: make([]*descriptorpb.FileDescriptorProto, len(files)),
	}

	for i, file := range files {
		fdp := protodesc.ToFileDescriptorProto(file)
		fds.File[i] = fdp

		if value, _ := protoregistry.GlobalFiles.FindFileByPath(fdp.GetName()); value != nil {
			zerolog.Ctx(ctx).Warn().
				Str("name", fdp.GetName()).
				Str("path", file.Path()).
				Msg("File already registered")

			continue
		}

		err := protoregistry.GlobalFiles.RegisterFile(file)
		if err != nil {
			return nil, errors.Wrapf(err, "error registering file %s", file.Path())
		}
	}

	return fds, nil
}

//nolint:cyclop
func compile(ctx context.Context, configure *Configure) ([]*descriptorpb.FileDescriptorSet, error) {
	capacity := len(configure.Descriptors())
	if len(configure.Protos()) > 0 {
		capacity++
	}

	results := make([]*descriptorpb.FileDescriptorSet, 0, capacity)

	for _, descriptor := range configure.Descriptors() {
		descriptorBytes, err := os.ReadFile(descriptor) //nolint:gosec
		if err != nil {
			return nil, errors.Wrapf(err, "failed to read descriptor: %s", descriptor)
		}

		fds := &descriptorpb.FileDescriptorSet{}

		err = proto.Unmarshal(descriptorBytes, fds)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal descriptor: %s", descriptor)
		}

		for _, fd := range fds.GetFile() {
			if value, _ := protoregistry.GlobalFiles.FindFileByPath(fd.GetName()); value != nil {
				zerolog.Ctx(ctx).Warn().
					Str("name", fd.GetName()).
					Str("path", descriptor).
					Msg("File already registered")

				continue
			}

			fileDesc, err := protodesc.NewFile(fd, protoregistry.GlobalFiles)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to create file descriptor: %s", descriptor)
			}

			err = protoregistry.GlobalFiles.RegisterFile(fileDesc)
			if err != nil {
				return nil, errors.Wrapf(err, "error registering file %s", descriptor)
			}
		}

		results = append(results, fds)
	}

	if len(configure.Protos()) > 0 {
		fds, err := createDescriptorSet(ctx, configure)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create descriptor set")
		}

		results = append(results, fds)
	}

	return results, nil
}

func newConfigure(ctx context.Context, imports []string, paths []string) (*Configure, error) {
	p := newProcessor(imports)

	err := p.process(ctx, paths)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create configuration")
	}

	return p.result(), nil
}

func findMinimalPaths(paths []string) []string {
	sort.Slice(paths, func(i, j int) bool {
		return len(paths[i]) < len(paths[j])
	})

	var result []string

	for _, path := range paths {
		isSubPath := false

		for _, existing := range result {
			rel, err := filepath.Rel(existing, path)
			if err != nil {
				continue
			}

			if !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
				isSubPath = true

				break
			}
		}

		if !isSubPath {
			result = append(result, path)
		}
	}

	return result
}

func Build(ctx context.Context, imports []string, paths []string) ([]*descriptorpb.FileDescriptorSet, error) {
	var err error

	for i, importPath := range imports {
		imports[i], err = filepath.Abs(importPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve import path: %s", importPath)
		}
	}

	for i, path := range paths {
		paths[i], err = filepath.Abs(path)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to resolve path: %s", path)
		}
	}

	configure, err := newConfigure(ctx, lo.Uniq(findMinimalPaths(imports)), lo.Uniq(paths))
	if err != nil {
		return nil, errors.Wrap(err, "create configuration")
	}

	return compile(ctx, configure)
}

type processor struct {
	imports          []string
	protos           []string
	descriptors      []string
	seenDirs         map[string]bool
	seenFiles        map[string]bool
	allowedProtoExts []string
	allowedDescExts  []string
}

func newProcessor(initialImports []string) *processor {
	return &processor{
		imports:   initialImports,
		seenDirs:  make(map[string]bool),
		seenFiles: make(map[string]bool),
		allowedProtoExts: []string{
			ProtoExt,
		},
		allowedDescExts: []string{
			ProtobufSetExt,
			ProtoSetExt,
		},
	}
}

func (p *processor) process(ctx context.Context, paths []string) error {
	logger := zerolog.Ctx(ctx)

	for _, path := range paths {
		select {
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck
		default:
		}

		logger.Debug().Str("path", path).Msg("Processing path")

		info, err := os.Stat(path)
		if err != nil {
			return errors.Wrapf(err, "failed to stat path: %s", path)
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return errors.Wrapf(err, "failed to resolve absolute path: %s", path)
		}

		switch {
		case info.IsDir():
			logger.Debug().Str("directory", absPath).Msg("Processing directory")

			err := p.processDirectory(ctx, absPath)
			if err != nil {
				return errors.Wrapf(err, "failed to process directory: %s", absPath)
			}
		default:
			logger.Debug().Str("file", absPath).Msg("Processing file")

			err := p.processFile(ctx, absPath)
			if err != nil {
				return errors.Wrapf(err, "failed to process file: %s", absPath)
			}
		}
	}

	return nil
}

func (p *processor) processDirectory(ctx context.Context, absPath string) error {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Str("directory", absPath).Msg("Walking directory")

	p.addImport(ctx, absPath)

	return errors.Wrapf(filepath.Walk(absPath, func(pth string, info os.FileInfo, err error) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return errors.Wrapf(err, "file system error at path: %s", pth)
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(pth)
		logger := logger.With().
			Str("file", pth).
			Str("extension", ext).
			Logger()

		switch {
		case slices.Contains(p.allowedProtoExts, ext):
			logger.Debug().Msg("Found proto file")
			p.addProtoFile(ctx, pth)
		case slices.Contains(p.allowedDescExts, ext):
			logger.Debug().Msg("Found descriptor file")
			p.addDescriptorFile(ctx, pth)
		default:
			logger.Debug().Msg("Skipping unsupported file type")
		}

		return nil
	}), "directory walk failed: %s", absPath)
}

func (p *processor) processFile(ctx context.Context, absPath string) error {
	logger := zerolog.Ctx(ctx).With().Str("file", absPath).Logger()

	ext := filepath.Ext(absPath)

	p.addImport(ctx, filepath.Dir(absPath))

	switch {
	case slices.Contains(p.allowedProtoExts, ext):
		logger.Debug().Msg("Adding proto file")
		p.addProtoFile(ctx, absPath)
	case slices.Contains(p.allowedDescExts, ext):
		logger.Debug().Msg("Adding descriptor file")
		p.addDescriptorFile(ctx, absPath)
	default:
		return errors.Wrapf(errUnsupportedFileType, "unsupported file: %s", absPath)
	}

	return nil
}

func (p *processor) addImport(ctx context.Context, dir string) {
	var (
		dirAbs string
		err    error
	)

	dirAbs, err = filepath.Abs(dir)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("dir", dir).Msg("Failed to resolve absolute path for import")

		return
	}

	if !p.seenDirs[dirAbs] {
		beforeLen := len(p.imports)

		p.imports = findMinimalPaths(append(p.imports, dirAbs))
		p.seenDirs[dirAbs] = true

		if len(p.imports) > beforeLen {
			zerolog.Ctx(ctx).Debug().Str("import", dirAbs).Msg("Added import path")
		}
	}
}

func (p *processor) addProtoFile(ctx context.Context, filePath string) {
	p.addFile(ctx, filePath, fileTypeProto)
}

func (p *processor) addDescriptorFile(ctx context.Context, filePath string) {
	p.addFile(ctx, filePath, fileTypeDescriptor)
}

func findPathByImports(filePath string, imports []string) (string, string) {
	filePath = filepath.ToSlash(filePath)

	sort.Slice(imports, func(i, j int) bool {
		return len(imports[i]) > len(imports[j])
	})

	for _, imp := range imports {
		impPath := filepath.ToSlash(imp)

		if !strings.HasSuffix(impPath, "/") {
			impPath += "/"
		}

		if strings.HasPrefix(filePath, impPath) {
			relPath := filePath[len(impPath):]

			return filepath.FromSlash(imp), filepath.FromSlash(relPath)
		}
	}

	return "", filepath.Base(filePath)
}

func (p *processor) addFile(ctx context.Context, filePath, fileType string) {
	var (
		fileAbs string
		err     error
	)

	fileAbs, err = filepath.Abs(filePath)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("file", filePath).Msg("Failed to resolve absolute path")

		return
	}

	if p.seenFiles[fileAbs] {
		zerolog.Ctx(ctx).Debug().Msg("File already processed")

		return
	}

	baseDir, _ := findPathByImports(fileAbs, p.imports)

	relPath, err := filepath.Rel(baseDir, fileAbs)
	if err != nil {
		zerolog.Ctx(ctx).Err(err).Str("file", fileAbs).Str("base_dir", baseDir).Msg("Failed to get relative path")

		return
	}

	switch fileType {
	case fileTypeProto:
		p.protos = append(p.protos, relPath)
	case fileTypeDescriptor:
		p.descriptors = append(p.descriptors, fileAbs)
	default:
		zerolog.Ctx(ctx).Error().Str("file_type", fileType).Msg("Unknown file type encountered")

		return
	}

	p.seenFiles[fileAbs] = true

	zerolog.Ctx(ctx).Debug().Str("type", fileType).Msg("File added successfully")
}

func (p *processor) result() *Configure {
	return &Configure{
		imports:     lo.Uniq(p.imports),
		protos:      lo.Uniq(p.protos),
		descriptors: lo.Uniq(p.descriptors),
	}
}

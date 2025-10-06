package runner

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	gofmt "go/format"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/BurntSushi/toml"
	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"

	"github.com/wasilibs/go-prettier/v3/internal/wasm"
)

var (
	errCheckFailed       = errors.New("check failed")
	errInvalidConfigFile = errors.New("invalid config file")
)

func NewRunner() *Runner {
	ctx := context.Background()

	rtCfg := wazero.NewRuntimeConfig()
	uc, err := os.UserCacheDir()
	if err == nil {
		cache, err := wazero.NewCompilationCacheWithDir(filepath.Join(uc, "com.github.wasilibs"))
		if err == nil {
			rtCfg = rtCfg.WithCompilationCache(cache)
		}
	}
	rt := wazero.NewRuntimeWithConfig(ctx, rtCfg)

	wasi_snapshot_preview1.MustInstantiate(ctx, rt)

	compiled, err := rt.CompileModule(ctx, wasm.Prettier)
	if err != nil {
		// Programming bug
		panic(err)
	}

	return &Runner{
		compiled: compiled,
		rt:       rt,
	}
}

type Runner struct {
	compiled wazero.CompiledModule
	rt       wazero.Runtime
}

type RunArgs struct {
	Cwd                       string
	Patterns                  []string
	Config                    string
	NoConfig                  bool
	NoEditorConfig            bool
	Check                     bool
	IgnorePaths               []string
	IgnoreUnknown             bool
	Write                     bool
	WithNodeModules           bool
	NoErrorOnUnmatchedPattern bool
}

func (r *Runner) Run(ctx context.Context, args RunArgs) error {
	var eCfg *editorconfig.Editorconfig

	// We use an untyped map for prettier config to allow piping through user config
	// without needing to recognizing every option.
	pCfg := map[string]any{}

	if !args.NoEditorConfig {
		eCfgPath := findConfigFile(args.Cwd, ".editorconfig")
		if eCfgPath != "" {
			f, err := os.Open(eCfgPath) //nolint:gosec
			// Ignore errors for best-effort features like editorconfig loading.
			if err == nil {
				if c, err := editorconfig.Parse(f); err == nil {
					eCfg = c
				}
			}
		}
	}

	switch {
	case args.Config != "":
		cfg, err := loadConfigFile(ctx, args.Config)
		if err != nil {
			return err
		}
		pCfg = cfg
	case args.NoConfig:
		// Do nothing
	default:
		for _, name := range []string{".prettierrc", ".prettierrc.json", ".prettierrc.yaml", ".prettierrc.yml", ".prettierrc.toml"} {
			if p := findConfigFile(args.Cwd, name); p != "" {
				cfg, err := loadConfigFile(ctx, p)
				if err != nil {
					return err
				}
				pCfg = cfg
				break
			}
		}
	}

	paths := expandPatterns(ctx, args)

	if args.Check {
		fmt.Println("Checking formatting...")
	}

	var numCheckFailed atomic.Uint32

	var g errgroup.Group
	for _, p := range paths {
		g.Go(func() error {
			if p.error != "" {
				slog.ErrorContext(ctx, p.error)
				return errors.New(p.error)
			}
			err := r.format(ctx, p, eCfg, pCfg, args.Check, args.Write, args.IgnoreUnknown)
			if errors.Is(err, errCheckFailed) {
				numCheckFailed.Add(1)
			}
			return err
		})
	}
	err := g.Wait()

	if args.Check {
		if n := numCheckFailed.Load(); n > 0 {
			slog.Warn(fmt.Sprintf("Code style issues found in %d files. Run Prettier to fix.", n))
		} else {
			fmt.Println("All matched files use Prettier code style!")
		}
	}

	return err
}

type jsonMsg struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

func (r *Runner) format(ctx context.Context, path expandedPath, eCfg *editorconfig.Editorconfig, userCfg map[string]any, check bool, write bool, ignoreUnknown bool) error {
	mergedCfg := map[string]any{}
	if eCfg != nil {
		def, err := eCfg.GetDefinitionForFilename(path.filePath)
		if err == nil {
			fillEditorConfig(def, mergedCfg)
		}
	}

	mergePrettierConfig(mergedCfg, userCfg, path.filePath)

	mergedCfg["filepath"] = path.filePath
	pCfgBytes, err := json.Marshal(mergedCfg)
	if err != nil {
		// Programming bug
		panic(err)
	}

	fi, err := os.Stat(path.filePath)
	if err != nil {
		slog.WarnContext(ctx, fmt.Sprintf(`Unable to read file "%s"`, path.filePath))
		slog.WarnContext(ctx, err.Error())
		return fmt.Errorf("runner: stat-ing file: %w", err)
	}

	in, err := os.ReadFile(path.filePath)
	if err != nil {
		slog.WarnContext(ctx, fmt.Sprintf(`Unable to read file "%s"`, path.filePath))
		slog.WarnContext(ctx, err.Error())
		return fmt.Errorf("runner: reading file: %w", err)
	}

	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()

	inMsg := jsonMsg{
		Name: "input",
		Body: string(in),
	}

	resChan := make(chan string, 1)

	go func() {
		defer func() {
			_ = stdinW.Close()
			_ = stdoutR.Close()
		}()
		stdinJW := json.NewEncoder(stdinW)
		if err := stdinJW.Encode(inMsg); err != nil {
			panic(fmt.Errorf("runner: encoding input message for prettier: %w", err))
		}
		br := bufio.NewReader(stdoutR)
		for {
			line, err := br.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				panic(err)
			}
			var msg jsonMsg
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				panic(fmt.Errorf("runner: unmarshaling message from prettier: %w", err))
			}
			switch msg.Name {
			case "gofmt-request":
				// TODO: Consider if err needs to be handled.
				formatted, err := gofmt.Source([]byte(msg.Body))
				if err != nil {
					// This should only apply to an embedded string, treat it as best-effort.
					formatted = []byte(msg.Body)
				}
				msg := jsonMsg{
					Name: "gofmt-response",
					Body: string(formatted),
				}
				if err := stdinJW.Encode(msg); err != nil {
					panic(fmt.Errorf("runner: encoding gofmt response message for prettier: %w", err))
				}
			case "result":
				resChan <- msg.Body
				return
			}
		}
	}()

	mCfg := wazero.NewModuleConfig().
		WithName("").
		WithSysNanosleep().
		WithSysNanotime().
		WithSysWalltime().
		WithRandSource(rand.Reader).
		WithArgs("prettier", string(pCfgBytes)).
		WithStdin(stdinR).
		WithStderr(stdoutW). // Use stderr for communication to avoid buffering challenges
		WithStdout(os.Stderr)

	_, err = r.rt.InstantiateModule(ctx, r.compiled, mCfg)
	if err != nil {
		if se, ok := err.(*sys.ExitError); ok { //nolint:errorlint
			if se.ExitCode() == 10 {
				if !ignoreUnknown && !path.ignoreUnknown {
					slog.WarnContext(ctx, fmt.Sprintf(`No parser could be inferred for file "%s".`, path.filePath))
				}
				return nil
			}
		}
		err = fmt.Errorf("runner: failed to run prettier [%s]: %w", path.filePath, err)
		fmt.Println(err)
		return err
	}

	res := <-resChan
	if write {
		if err := os.WriteFile(path.filePath, []byte(res), fi.Mode()); err != nil {
			return fmt.Errorf("runner: failed to write file: %w", err)
		}
	} else if !check {
		fmt.Print(res)
	}

	if check && !bytes.Equal(in, []byte(res)) {
		slog.Warn(path.filePath)
		return errCheckFailed
	}

	return nil
}

func findConfigFile(cwd string, name string) string {
	dir, err := filepath.Abs(cwd)
	if err != nil {
		return ""
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			return filepath.Join(dir, name)
		}

		parent := filepath.Dir(dir)
		if parent == dir || parent == "" {
			return ""
		}

		dir = parent
	}
}

func loadConfigFile(ctx context.Context, path string) (map[string]any, error) {
	res := map[string]any{}

	pCfgBytes, err := os.ReadFile(path) //nolint:gosec
	if err != nil {
		slog.WarnContext(ctx, fmt.Sprintf(`Unable to read config file "%s"`, path))
		slog.WarnContext(ctx, err.Error())
		return res, fmt.Errorf("runner: reading config file: %w", err)
	}

	// YAML is superset of JSON so it should be fine to only use YAML to parse.
	err = yaml.Unmarshal(pCfgBytes, &res)
	if err == nil {
		return res, nil
	}

	if err := toml.Unmarshal(pCfgBytes, &res); err == nil {
		return res, nil
	}

	slog.WarnContext(ctx, fmt.Sprintf(`Invalid config file "%s"`, path))
	// JSON / YAML are more common so use it's error rather than TOML's
	slog.WarnContext(ctx, err.Error())
	return res, errInvalidConfigFile
}

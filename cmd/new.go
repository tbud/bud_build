package cmd

import (
	"bytes"
	. "github.com/tbud/bud/common"
	"go/build"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var cmdNew = &Command{
	UsageLine: "new [-s] [path]",
	Short:     "create a bud application from seed",
	Long: `
New creates a few files to get a new bud application running quickly.

It puts all of the files in the given path, taking the final element in
the path to be the app name.

The -s flag is an optional argument, provided the ability to create from a special seed.
The default seed is react.

For example:

    bud new import/path/appname

    bud new -s react import/path/appname
    `,
}

var seedName = cmdNew.Flag.String("s", "tea", "")

var (
	srcRoot    string
	appPath    string
	appName    string
	basePath   string
	importPath string
)

func init() {
	cmdNew.Run = newCommand
	rand.Seed(time.Now().UnixNano())
}

func newCommand(cmd *Command, args []string) {
	if len(args) != 1 {
		LogFatalExit("Command error. Run 'bud help new' for usage.\n")
	}

	checkGoPaths()

	// checking and setting application
	parseParams(args)

	// checking and copy from seed
	copyFromSeed()

	Log.Info("Your application is ready:\n  %s", appPath)
	Log.Info("\nYou can run it with:\n  bud run %s", importPath)
}

const alphaNumeric = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func generateSecret() string {
	chars := make([]byte, 64)
	for i := 0; i < 64; i++ {
		chars[i] = alphaNumeric[rand.Intn(len(alphaNumeric))]
	}

	return string(chars)
}

// lookup and set Go related variables
func checkGoPaths() {
	// lookup go path
	gopath := build.Default.GOPATH
	if gopath == "" {
		LogFatalExit("Abort: GOPATH environment variable is not set. " +
			"Please refer to http://golang.org/doc/code.html to configure your Go environment.")
	}

	// set go src path
	srcRoot = filepath.Join(filepath.SplitList(gopath)[0], "src")
}

func parseParams(args []string) {
	var err error
	importPath = args[0]
	if filepath.IsAbs(importPath) {
		LogFatalExit("Abort: '%s' looks like a directory.  Please provide a Go import path instead.",
			importPath)
	}

	_, err = build.Import(importPath, "", build.FindOnly)
	if err == nil {
		LogFatalExit("Abort: Import path %s already exists.\n", importPath)
	}

	appPath = filepath.Join(srcRoot, filepath.FromSlash(importPath))
	appName = filepath.Base(appPath)
	basePath = filepath.ToSlash(filepath.Dir(importPath))

	if basePath == "." {
		// we need to remove the a single '.' when
		// the app is in the $GOROOT/src directory
		basePath = ""
	} else {
		// we need to append a '/' when the app is
		// is a subdirectory such as $GOROOT/src/path/to/revelapp
		basePath += "/"
	}
}

func copyFromSeed() {
	var seedPath string
	if strings.Index(*seedName, string(os.PathSeparator)) == -1 {
		checkAndGetImport(BUD_DEFAULT_SEED_PATH)

		budSeedPath, err := build.Import(BUD_DEFAULT_SEED_PATH, "", build.FindOnly)
		panicOnError(err, "import bud seed path error:%s", BUD_DEFAULT_SEED_PATH)

		seedPath = filepath.Join(budSeedPath.Dir, *seedName)
	} else {
		checkAndGetImport(*seedName)
		seedPath = filepath.Join(srcRoot, *seedName)
	}

	// copy files to new app directory
	copySeedArchetype(appPath, seedPath, map[string]interface{}{
		"AppName":  appName,
		"BasePath": basePath,
		"Secret":   generateSecret(),
	})

}

func checkAndGetImport(path string) {
	gocmd, err := exec.LookPath("go")
	if err != nil {
		LogFatalExit("Go executable not found in PATH.")
	}

	_, err = build.Import(path, "", build.FindOnly)
	if err != nil {
		getCmd := exec.Command(gocmd, "get", "-d", path)
		Log.Info("Exec: %s", getCmd.Args)
		bOutput, err := getCmd.CombinedOutput()

		bpos := bytes.Index(bOutput, []byte("no buildable Go source files in"))
		if err != nil && bpos == -1 {
			LogFatalExit("Abort: Could not find or 'go get' path '%s'.\nOutput: %s", path, bOutput)
		}
	}
}

func copySeedArchetype(destDir, srcDir string, data map[string]interface{}) error {
	var originSrcDir string
	// check seed dir wether or not a link
	fi, err := os.Lstat(srcDir)
	if err == nil && fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		originSrcDir, err = os.Readlink(srcDir)
		panicOnError(err, "Read link err %s", srcDir)
	} else {
		originSrcDir = srcDir
	}

	// check seed archetype dir is exist
	archetypeDir := filepath.Join(originSrcDir, "archetype")
	if _, err := os.Stat(archetypeDir); err != nil {
		if os.IsNotExist(err) {
			LogFatalExit("Seed archetype not exist: %s", archetypeDir)
		}
	}

	err = os.MkdirAll(appPath, 0777)
	panicOnError(err, "Failed to create directory: %s", appPath)

	return filepath.Walk(archetypeDir, func(path string, info os.FileInfo, err error) error {
		relSrcPath := strings.TrimLeft(path[len(archetypeDir):], string(os.PathSeparator))
		destPath := filepath.Join(destDir, relSrcPath)

		if strings.HasPrefix(relSrcPath, ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			err = os.MkdirAll(destPath, 0777)
			if !os.IsNotExist(err) {
				panicOnError(err, "Failed to create directory: %s", destPath)
			}
			return nil
		}

		if strings.HasSuffix(relSrcPath, ".template") {
			copyTemplateFile(destPath[:len(destPath)-len(".template")], path, data)
			return nil
		}

		copyFile(destPath, path)
		return nil
	})
}

func copyFile(destFile, srcFile string) {
	dst, err := os.Create(destFile)
	panicOnError(err, "Failed to create file: %s, %s", destFile, err)

	src, err := os.Open(srcFile)
	panicOnError(err, "Failed to open file: %s, %s", srcFile, err)

	_, err = io.Copy(dst, src)
	panicOnError(err, "Failed to copy data from %s to %s with err: %s", dst.Name(), src.Name(), err)

	panicOnError(src.Close(), "Failed to close file %s", src.Name())

	panicOnError(dst.Close(), "Failed to close file %s", dst.Name())
}

func copyTemplateFile(destFile, srcFile string, data map[string]interface{}) {
	temp, err := template.ParseFiles(srcFile)
	panicOnError(err, "Failed to parse template %s", srcFile)

	dst, err := os.Create(destFile)
	panicOnError(err, "Failed to create file %s", dst.Name())

	panicOnError(temp.Execute(dst, data), "Failed to render template %s", srcFile)

	panicOnError(dst.Close(), "Failed to close file %s", dst.Name())
}

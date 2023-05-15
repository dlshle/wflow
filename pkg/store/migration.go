package store

import (
	"bufio"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/dlshle/gommon/utils"
	"github.com/jmoiron/sqlx"
)

func ExecMigration(db *sqlx.DB, scriptPath string) error {
	scripts, err := loadMigrationScripts(scriptPath)
	if err != nil {
		return err
	}
	return ExecMigrationScripts(db, scripts)
}

func loadMigrationScripts(scriptDir string) (scriptsByVersions []string, err error) {
	var (
		dirFile *os.File
		entries []fs.DirEntry
	)
	defer func() {
		dirFile.Close()
	}()
	scriptsByVersions = make([]string, 0)
	err = utils.ProcessWithErrors(func() error {
		dirFile, err = os.Open(scriptDir) // For read access.
		return err
	}, func() error {
		entries, err = dirFile.ReadDir(0)
		return err
	}, func() error {
		scriptPaths := getScriptPathsFromDirEntriesBySuffix(scriptDir, entries)
		sort.Strings(scriptPaths)
		for _, entry := range scriptPaths {
			script, err := readTextFile(entry)
			if err != nil {
				return err
			}
			scriptsByVersions = append(scriptsByVersions, script)
		}
		return nil
	})
	return
}

func getScriptPathsFromDirEntriesBySuffix(scriptDir string, entries []fs.DirEntry) []string {
	var nonDirEntries []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			nonDirEntries = append(nonDirEntries, scriptDir+"/"+entry.Name())
		}
	}
	return nonDirEntries
}

func readTextFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var builder strings.Builder
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		builder.WriteString(line)
	}
	return builder.String(), nil
}

func ExecMigrationScripts(db *sqlx.DB, migrationScripts []string) error {
	for _, migration := range migrationScripts {
		_, err := db.Exec(migration)
		if err != nil {
			return err
		}
	}
	return nil
}

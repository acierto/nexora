package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func initialize(root, project string) error { return initializeWith(root, project, writeExclusive) }

// initializeWith permits deterministic failure testing of cleanup after creation.
func initializeWith(root, project string, write func(string, []byte) error) error {
	if strings.TrimSpace(project) == "" {
		return fmt.Errorf("project: must not be empty")
	}
	config := []byte("schema_version: 1\nproject: " + strconv.Quote(project) + "\nsources:\n  - id: local\n    path: ./knowledge\n")
	rule := []byte("---\nid: project/conventions\nname: Project conventions\ndescription: Development conventions for this project\nkind: rule\napplies_to:\n  projects: [" + strconv.Quote(project) + "]\n---\nKeep changes focused and test the behavior you change.\n")
	paths := []string{"nexora.yaml", filepath.Join("knowledge", "rules", "project.md")}
	for _, rel := range paths {
		parts := strings.Split(rel, string(filepath.Separator))
		current := root
		for idx, part := range parts {
			current = filepath.Join(current, part)
			info, err := os.Lstat(current)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("init refuses symlink destination %s", current)
			}
			if idx == len(parts)-1 {
				return fmt.Errorf("init target already exists: %s", current)
			}
			if !info.IsDir() {
				return fmt.Errorf("init destination component is not a directory: %s", current)
			}
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "knowledge", "rules"), 0755); err != nil {
		return err
	}
	created := []string{}
	for idx, data := range [][]byte{config, rule} {
		file := filepath.Join(root, paths[idx])
		if err := write(file, data); err != nil {
			for _, owned := range created {
				err = errors.Join(err, os.Remove(owned))
			}
			return fmt.Errorf("init: %w", err)
		}
		created = append(created, file)
	}
	return nil
}

func writeExclusive(file string, data []byte) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	_, writeErr := f.Write(data)
	closeErr := f.Close()
	err = errors.Join(writeErr, closeErr)
	if err != nil {
		return errors.Join(err, os.Remove(file))
	}
	return nil
}

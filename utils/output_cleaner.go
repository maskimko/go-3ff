package utils

import (
	"log"
	"os"
	"path/filepath"
)

func ResetOutput(dir string, environments, layers []string) error {
	err := DeleteOutput(dir, environments, layers)
	if err != nil {
		return err
	}
	if _, e := os.Stat(dir); os.IsNotExist(e) {
		err = os.MkdirAll(filepath.Join(dir, "generator", "output"), 0755)
	}
	return err
}

func DeleteOutput(dir string, environments, layers []string) error {
	for _, env := range environments {
		ed := filepath.Join(dir, "generator", "output", env)
		for _, layer := range layers {
			ld := filepath.Join(ed, layer)
			err := delete(ld)
			if err != nil {
				return err
			}
		}
		if len(layers) == 0 {
			err := delete(ed)
			return err
		}
	}
	return nil
}

func delete(dir string) error {
	if _, e := os.Stat(dir); !os.IsNotExist(e) {
		err := os.RemoveAll(dir)
		if err != nil {
			if Debug {
				log.Printf("Cannot delete directory %s Error: %s", dir, err)
			}
		}
		return err
	}
	return nil
}

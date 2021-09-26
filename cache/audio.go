package cache

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	"github.com/bcurnow/magicband-reader/config"
	"github.com/bcurnow/magicband-reader/context"
	"github.com/bcurnow/magicband-reader/rfidsecuritysvc"
)

func init() {
	sounds, err := context.RFIDSecuritySvc.Sounds().List()
	if err != nil {
		panic(err)
	}
	mappedSounds := mapSounds(sounds)

	files, err := os.ReadDir(config.SoundDir)
	if err != nil {
		panic(err)
	}
	mappedFiles := mapFiles(files)

	for _, sound := range mappedSounds {
		if dirInfo, exists := mappedFiles[sound.Name]; exists {
			// check the timestamps
			fsInfo, err := dirInfo.Info()
			if err != nil {
				panic(err)
			}

			if fsInfo.ModTime().Before(sound.LastUpdateTimestamp) {
				// download
				fmt.Println(fmt.Sprintf("Local file found but out of date, downloading %v", sound.Name))
				if err := writeSound(sound.Name); err != nil {
					panic(err)
				}
			}
		} else {
			// download the file
			fmt.Println(fmt.Sprintf("Downloading %v", sound.Name))
			if err := writeSound(sound.Name); err != nil {
				panic(err)
			}
		}
	}
}

func mapSounds(sounds []rfidsecuritysvc.Sound) map[string]rfidsecuritysvc.Sound {
	results := make(map[string]rfidsecuritysvc.Sound)
	for _, s := range sounds {
		results[s.Name] = s
	}
	return results
}

func mapFiles(files []os.DirEntry) map[string]os.DirEntry {
	results := make(map[string]os.DirEntry)
	for _, f := range files {
		results[f.Name()] = f
	}
	return results
}

func writeSound(name string) error {
	sound, err := context.RFIDSecuritySvc.Sounds().Get(name)
	if err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(sound.Content)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(config.SoundDir, sound.Name), data, 0660); err != nil {
		return err
	}
	return nil
}

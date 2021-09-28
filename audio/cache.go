package audio

import (
	"encoding/base64"
	"fmt"
	"os"
	"path"

	log "github.com/sirupsen/logrus"

	"github.com/bcurnow/magicband-reader/rfidsecuritysvc"
)

type Cache interface {
	Sync() error
	Load(name string) (*os.File, error)
	CacheDir() string
}

type cache struct {
	rfidSecuritySvc rfidsecuritysvc.Service
	soundDir        string
}

func NewCache(svc rfidsecuritysvc.Service, soundDir string) (Cache, error) {
	log.Trace("Creating new audio.Controller")
	return &cache{rfidSecuritySvc: svc, soundDir: soundDir}, nil
}

func (c *cache) Sync() error {
	sounds, err := c.rfidSecuritySvc.Sounds().List()
	if err != nil {
		return err
	}
	mappedSounds := mapSounds(sounds)

	files, err := os.ReadDir(c.soundDir)
	if err != nil {
		return err
	}
	mappedFiles := mapFiles(files)

	for _, sound := range mappedSounds {
		if dirInfo, exists := mappedFiles[sound.Name]; exists {
			// check the timestamps
			fsInfo, err := dirInfo.Info()
			if err != nil {
				return err
			}

			if fsInfo.ModTime().Before(sound.LastUpdateTimestamp) {
				log.Debugf("%v found but out of date, updating", dirInfo.Name())
				if err := c.writeSound(sound.Name); err != nil {
					return err
				}
			}
		} else {
			log.Debugf("%v not found, downloading to %v", sound.Name, c.soundDir)
			if err := c.writeSound(sound.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *cache) Load(name string) (*os.File, error) {
	filePath := path.Join(c.soundDir, name)
	f, err := os.Open(filePath)
	if err != nil {
		// The file does not exist, write it to the directory and re-open
		if err := c.writeSound(name); err != nil {
			f, err := os.Open(filePath)
			if err != nil {
				// We though we downloaded but something went wrong
				return nil, fmt.Errorf("%v not found in cache, downloading was successful but still unable to read %v: %v", name, filePath, err)
			}
			return f, nil
		}
	}
	return f, nil
}

func (c *cache) CacheDir() string {
	return c.soundDir
}

func (c *cache) writeSound(name string) error {
	sound, err := c.rfidSecuritySvc.Sounds().Get(name)
	if err != nil {
		return err
	}

	data, err := base64.StdEncoding.DecodeString(sound.Content)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path.Join(c.soundDir, sound.Name), data, 0660); err != nil {
		return err
	}
	return nil
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

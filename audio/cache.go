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
	CacheDir() string
	Get(soundName string) (*os.File, error)
	Load(sound *rfidsecuritysvc.Sound) (*os.File, error)
	Sync() error
}

type cache struct {
	rfidSecuritySvc rfidsecuritysvc.Service
	soundDir        string
}

func NewCache(svc rfidsecuritysvc.Service, soundDir string) (Cache, error) {
	log.Trace("Creating new audio.Controller")
	return &cache{rfidSecuritySvc: svc, soundDir: soundDir}, nil
}

func (c *cache) CacheDir() string {
	return c.soundDir
}

func (c *cache) Get(soundName string) (*os.File, error) {
	filePath := path.Join(c.soundDir, soundName)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return f, err
}

func (c *cache) Load(sound *rfidsecuritysvc.Sound) (*os.File, error) {
	f, err := c.Get(sound.Name)
	if err == nil {
		return f, nil
	}

	// File doesn't currently existing on the file system, download from the service
	if err := c.writeSound(sound.ID); err != nil {
		return nil, err
	}

	f, err = c.Get(sound.Name)
	if err != nil {
		// We though we downloaded but something went wrong
		return nil, fmt.Errorf("%v not found in cache, downloading was successful but still unable to read: %v", sound.Name, err)
	}
	return f, nil
}

func (c *cache) Sync() error {
	mappedSounds, err := c.mapSounds()
	if err != nil {
		return err
	}

	mappedFiles, err := c.mapFiles()
	if err != nil {
		return err
	}

	for _, sound := range mappedSounds {
		if dirInfo, exists := mappedFiles[sound.Name]; exists {
			// check the timestamps
			fsInfo, err := dirInfo.Info()
			if err != nil {
				return err
			}

			if fsInfo.ModTime().Before(sound.LastUpdateTimestamp) {
				log.Debugf("%v found but out of date, updating", dirInfo.Name())
				if err := c.writeSound(sound.ID); err != nil {
					return err
				}
			}
		} else {
			log.Debugf("%v not found, downloading to %v", sound.Name, c.soundDir)
			if err := c.writeSound(sound.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *cache) writeSound(id int) error {
	sound, err := c.rfidSecuritySvc.Sounds().Get(id)
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

func (c *cache) mapSounds() (map[string]rfidsecuritysvc.Sound, error) {
	sounds, err := c.rfidSecuritySvc.Sounds().List()
	if err != nil {
		return nil, err
	}

	results := make(map[string]rfidsecuritysvc.Sound)
	for _, s := range sounds {
		results[s.Name] = s
	}
	return results, nil
}

func (c *cache) mapFiles() (map[string]os.DirEntry, error) {
	files, err := os.ReadDir(c.soundDir)
	if err != nil {
		return nil, err
	}

	results := make(map[string]os.DirEntry)
	for _, f := range files {
		results[f.Name()] = f
	}
	return results, nil
}

package discinfo

import (
	"fmt"
	"github.com/b0bbywan/go-cd-cuer/types"
	"github.com/b0bbywan/go-cd-cuer/utils"
	"log"
	"os"
	"path/filepath"
)

func GenerateCueFile(info *types.DiscInfo, cueFilePath string) error {
	file, err := os.Create(cueFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	if info.CoverArtPath == "" {
		coverFilePath := utils.CacheCoverArtPath(filepath.Base(filepath.Dir(cueFilePath)))
		if err := fetchCoverArt(info.ID, coverFilePath); err == nil {
			info.CoverArtPath = coverFilePath
		} else {
			log.Printf("error getting cover: %v", err)
		}
	}

	var content string
	if info.ReleaseDate != "" {
		content += fmt.Sprintf("REM DATE \"%s\"\n", info.ReleaseDate)
	}
	if info.Genre != "" {
		content += fmt.Sprintf("REM GENRE \"%s\"\n", info.Genre)
	}
	if info.CoverArtPath != "" {
		content += fmt.Sprintf("REM COVER \"%s\"\n", info.CoverArtPath)
	}
	content += fmt.Sprintf("PERFORMER \"%s\"\nTITLE \"%s\"\n", info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}

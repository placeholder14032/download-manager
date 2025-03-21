package download

import (
    "fmt"
    "io"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

type PartsCombiner struct {
    BufferSize int // Size of the copy buffer - 32 kb default (fixed for now)
    ContentLength int64// Total size of the file
    PartsCount int 
	ChunkSize     int64
}
func NewPartsCombiner(contentLength int64, partsCount int, chunkSize int64) *PartsCombiner {
    return &PartsCombiner{
        BufferSize:    32 * 1024,
        ContentLength: contentLength,
        PartsCount:    partsCount,
        ChunkSize:     chunkSize,
    }
}

func (c *PartsCombiner) CombineParts(filePath string, contentLength int64, partsCount int) error {
	fmt.Println("Starting combine parts")
	
	// if it's already completed we don't need to do anything
    if c.isFileComplete(filePath, contentLength) {
        return nil
    }

	// finding file parts and making sure every part is exisiting
    partFiles, err := c.findPartFiles(filePath)
    if err != nil {
        return err
    }
    if len(partFiles) == 0 {
        return fmt.Errorf("no part files found to combine")
    }
	
	// mapping parts and their index stuff
    partsMap, err := c.parsePartFiles(filePath, partFiles)
    if err != nil {
        return err
    }
	// Verify all parts exist	
	if err := c.verifyParts(partsMap, partsCount); err != nil {
		return err
	}

	// merging parts
    if err := c.mergeParts(filePath, partsMap); err != nil {
        return err
    }

	// making sure if merged size match the contentLength we expected
    if err := c.verifyCombinedFile(filePath, contentLength); err != nil {
        return err
    }

	// cleaning up part files we don't need anymore
    c.cleanupPartFiles(partFiles)
	fmt.Println("Combine parts completed successfully")
    return nil
}

func (c *PartsCombiner) isFileComplete(filePath string, contentLength int64) bool {
    info, err := os.Stat(filePath)
    return err == nil && info.Size() == contentLength
}

func (c *PartsCombiner) findPartFiles(filePath string) ([]string, error) {
    partFiles, err := filepath.Glob(fmt.Sprintf("%s.part*", filePath))
    if err != nil {
        return nil, fmt.Errorf("failed to find part files: %v", err)
    }
    return partFiles, nil
}

func (c *PartsCombiner) parsePartFiles(basePath string, partFiles []string) (map[int]string, error) {
    baseName := filepath.Base(basePath)
    partsMap := make(map[int]string)
    for _, partFile := range partFiles {
        partBase := filepath.Base(partFile)
        if !strings.HasPrefix(partBase, baseName+".part") {
            continue
        }
        numStr := strings.TrimPrefix(partBase, baseName+".part")
        partNum, err := strconv.Atoi(numStr)
        if err != nil {
            return nil, fmt.Errorf("invalid part file name %s: %v", partFile, err)
        }
        partsMap[partNum] = partFile
    }
    return partsMap, nil
}

func (c *PartsCombiner) verifyParts(partsMap map[int]string, partsCount int) error {
    for i := 0; i < partsCount; i++ {
        if _, exists := partsMap[i]; !exists {
            return fmt.Errorf("missing part file %d", i)
        }
    }
    return nil
}

func (c *PartsCombiner) mergeParts(filePath string, partsMap map[int]string) error {
    fmt.Println("Merging parts...")
    // combinedFile, err := os.Create(filePath)
	combinedFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644) // with this we can overwrite on that
    if err != nil {
        return fmt.Errorf("failed to create final file: %v", err)
    }
    defer combinedFile.Close()

    buffer := make([]byte, c.BufferSize)
    totalWritten := int64(0)

    for i := 0; i < c.PartsCount; i++ {
        partFilePath := partsMap[i]
        partFile, err := os.Open(partFilePath)
        if err != nil {
            return fmt.Errorf("failed to open part %d: %v", i, err)
        }
        defer partFile.Close()

        info, err := partFile.Stat()
        if err != nil {
            return fmt.Errorf("failed to stat part %d: %v", i, err)
        }
        partSize := info.Size()
        if partSize == 0 {
            return fmt.Errorf("part %d is empty", i)
        }

        expectedSize := c.ChunkSize
        if i == c.PartsCount-1 {
            remaining := c.ContentLength % c.ChunkSize
            if remaining != 0 {
                expectedSize = remaining
            }
        }

        if partSize != expectedSize {
            return fmt.Errorf("part %d size mismatch: got %d, want %d", i, partSize, expectedSize)
        }

        written, err := io.CopyBuffer(combinedFile, partFile, buffer)
        if err != nil {
            return fmt.Errorf("failed to copy part %d: %v", i, err)
        }
        if written != partSize {
            return fmt.Errorf("part %d copy mismatch: wrote %d, expected %d", i, written, partSize)
        }
        fmt.Printf("Combined part %d: %d bytes\n", i, written)
        totalWritten += written
    }
    fmt.Printf("Total bytes written: %d\n", totalWritten)
    return nil
}

func (c *PartsCombiner) verifyCombinedFile(filePath string, contentLength int64) error {
    info, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("failed to verify final file: %v", err)
    }
    if info.Size() != contentLength {
        return fmt.Errorf("final file size mismatch: got %d, want %d", info.Size(), contentLength)
    }
    return nil
}

func (c *PartsCombiner) cleanupPartFiles(partFiles []string) {
    for _, partFile := range partFiles {
        if err := os.Remove(partFile); err != nil {
            fmt.Printf("Warning: failed to remove part file %s: %v\n", partFile, err)
        }
    }
}
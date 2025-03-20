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
}
func NewPartsCombiner() *PartsCombiner {
    return &PartsCombiner{BufferSize: 32 * 1024} 
}

func (c *PartsCombiner) CombineParts(filePath string, contentLength int64, partsCount int) error {
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

	// merging parts
    if err := c.mergeParts(filePath, partsMap, partsCount); err != nil {
        return err
    }

	// making sure if merged size match the contentLength we expected
    if err := c.verifyCombinedFile(filePath, contentLength); err != nil {
        return err
    }

	// cleaning up part files we don't need anymore
    c.cleanupPartFiles(partFiles)
    return nil
}

func (c *PartsCombiner) isFileComplete(filePath string, contentLength int64) bool {
    info, err := os.Stat(filePath)
    return err == nil && info.Size() == int64(contentLength)
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

func (c *PartsCombiner) mergeParts(filePath string, partsMap map[int]string, partsCount int) error {
    combinedFile, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create final file: %v", err)
    }
    defer combinedFile.Close()

    buffer := make([]byte, c.BufferSize)
    totalWritten := int64(0)

    for i := 0; i < partsCount; i++ {
        partFilePath := partsMap[i]
        partFile, err := os.Open(partFilePath)
        if err != nil {
            return fmt.Errorf("failed to open part %d: %v", i, err)
        }
        info, err := partFile.Stat()
        if err != nil {
            partFile.Close()
            return fmt.Errorf("failed to stat part %d: %v", i, err)
        }
        partSize := info.Size()
        if partSize == 0 {
            partFile.Close()
            return fmt.Errorf("part %d is empty", i)
        }
        expectedSize := int64(10240)

		// for last part we might have diffrent expevted size
        if i == partsCount-1 { 
            expectedSize = int64( c.ContentLength % 10240)
            if expectedSize == 0 {
                expectedSize = 10240
            }
        }

        if partSize != expectedSize {
            partFile.Close()
            return fmt.Errorf("part %d size mismatch before copy: got %d, want %d", i, partSize, expectedSize)
        }

        written, err := io.CopyBuffer(combinedFile, partFile, buffer)
        partFile.Close()
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

// just ensuring written size matches the content length we expected
func (h *PartsCombiner) verifyCombinedFile(filePath string, contentLength int64) error {
    info, err := os.Stat(filePath)
    if err != nil {
        return fmt.Errorf("failed to verify final file: %v", err)
    }
    if info.Size() != int64(contentLength) {
        return fmt.Errorf("final file size mismatch: got %d, want %d", info.Size(), contentLength)
    }
    return nil
}

func (h *PartsCombiner) cleanupPartFiles(partFiles []string) {
    for _, partFile := range partFiles {
        if err := os.Remove(partFile); err != nil {
            fmt.Printf("Warning: failed to remove part file %s: %v\n", partFile, err)
        }
    }
}
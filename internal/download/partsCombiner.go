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
    BufferSize int // Size of the copy buffer 
}
func NewPartsCombiner() *PartsCombiner {
    return &PartsCombiner{BufferSize: 32 * 1024} 
}

func (c *PartsCombiner) CombineParts(filePath string, contentLength, partsCount int) error {
    if c.isFileComplete(filePath, contentLength) {
        return nil
    }

    partFiles, err := c.findPartFiles(filePath)
    if err != nil {
        return err
    }
    if len(partFiles) == 0 {
        return fmt.Errorf("no part files found to combine")
    }

    partsMap, err := c.parsePartFiles(filePath, partFiles)
    if err != nil {
        return err
    }

    if err := c.verifyParts(partsMap, partsCount); err != nil {
        return err
    }

    if err := c.mergeParts(filePath, partsMap, partsCount); err != nil {
        return err
    }

    if err := c.verifyCombinedFile(filePath, contentLength); err != nil {
        return err
    }

    c.cleanupPartFiles(partFiles)
    return nil
}

func (c *PartsCombiner) isFileComplete(filePath string, contentLength int) bool {
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
    for i := 0; i < partsCount; i++ {
        partFile, err := os.Open(partsMap[i])
        if err != nil {
            return fmt.Errorf("failed to open part %d: %v", i, err)
        }
        _, err = io.CopyBuffer(combinedFile, partFile, buffer)
        partFile.Close()
        if err != nil {
            return fmt.Errorf("failed to copy part %d: %v", i, err)
        }
    }
    return nil
}

func (h *PartsCombiner) verifyCombinedFile(filePath string, contentLength int) error {
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
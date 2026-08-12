package sstable

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ccoveille/go-safecast/v2"
)

type Footer struct {
	Magic       uint64
	IndexOffset uint64
	//BloomOffset uint64
}

type IndexBlock struct {
	Key    string
	Offset uint64
}

type SSTFile struct {
	File       *os.File
	Generation int
	FileName   string
	Size       int64
	Index      []*IndexBlock
	Footer     *Footer

	//bloom filter goes here
}

func (s *SSTFile) Open(filepath string) error {
	if s.File != nil {
		return fmt.Errorf("file is already open")
	}

	s.FileName = filepath
	file, err := os.OpenFile(s.FileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)

	if err != nil {
		return err
	}

	s.File = file
	s.Generation, err = s.ParseGeneration()

	if err != nil {
		return err
	}

	return nil
}

func (s *SSTFile) Close() error {
	if s.File == nil {
		return nil
	}

	err := s.File.Close()
	s.File = nil

	return err
}

func (s *SSTFile) PopulateIndex(blocks []byte) {
	indexBlocks := make([]*IndexBlock, 0)

	for i := 0; i < len(blocks); {
		iU32, err := safecast.Convert[uint32](i)

		if err != nil {
			return 
		}
		keySize := binary.BigEndian.Uint32(blocks[i : i+4])

		keySizeInt, err := safecast.Convert[int](keySize)
		if err != nil {
			break
		}

		if i+12+keySizeInt > len(blocks) {
			break
		}

		indexKey := string(blocks[i+4 : iU32+4+keySize])

		keyOffset := binary.BigEndian.Uint64(blocks[iU32+4+keySize : iU32+12+keySize])
		//fmt.Println(keyOffset)
		index := &IndexBlock{
			Key:    indexKey,
			Offset: keyOffset,
		}

		i += keySizeInt + 12
		indexBlocks = append(indexBlocks, index)
	}

	s.Index = indexBlocks
}

func (s *SSTFile) PopulateFooter() error {
	footer := make([]byte, 24)
	_, err := s.File.ReadAt(footer, s.Size-24)

	if err != nil {
		return err
	}

	s.Footer = &Footer{}

	s.Footer.Magic = binary.BigEndian.Uint64(footer[16:])
	s.Footer.IndexOffset = binary.BigEndian.Uint64(footer[8:16])

	return nil
}

func (s *SSTFile) Delete() error {
	err := s.File.Close()

	if err != nil {
		return err
	}

	err = os.Remove(s.FileName)

	return err
}

func (s *SSTFile) ParseGeneration() (int, error) {
	filename := filepath.Base(s.FileName)

	str := strings.Split(filename, "_")[1]
	str = strings.Split(str, ".")[0]
	idx, err := strconv.Atoi(str)
	//s.Generation = idx
	return idx, err
}

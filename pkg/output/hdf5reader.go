package output

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/batchatco/go-native-netcdf/netcdf/api"
	"github.com/batchatco/go-native-netcdf/netcdf/hdf5"
)

// ExtractRadarGrid reads a MeteoSwiss ODIM_H5 radar file and extracts the precipitation grid.
// Pure Go implementation — no Python dependency.
func ExtractRadarGrid(h5path string) (*RadarGrid, error) {
	// Step 1: Read attributes using go-native-netcdf
	group, err := hdf5.Open(h5path)
	if err != nil {
		return nil, fmt.Errorf("open HDF5: %w", err)
	}
	defer group.Close()

	where, err := group.GetGroup("where")
	if err != nil {
		return nil, fmt.Errorf("get 'where' group: %w", err)
	}

	attrs := where.Attributes()
	rows := getAttrInt(attrs, "ysize")
	cols := getAttrInt(attrs, "xsize")
	if rows == 0 || cols == 0 {
		return nil, fmt.Errorf("invalid grid dimensions: %dx%d", rows, cols)
	}

	llLat := getAttrFloat(attrs, "LL_lat")
	llLon := getAttrFloat(attrs, "LL_lon")
	urLat := getAttrFloat(attrs, "UR_lat")
	urLon := getAttrFloat(attrs, "UR_lon")

	// Step 2: Read the raw file and find the compressed data chunk
	raw, err := os.ReadFile(h5path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	expectedSize := rows * cols * 8 // float64 = 8 bytes
	data, err := findAndDecompressChunk(raw, expectedSize)
	if err != nil {
		return nil, fmt.Errorf("extract data chunk: %w", err)
	}

	// Step 3: Unshuffle (HDF5 byte shuffle for 8-byte float64 elements)
	unshuffled := unshuffleBytes(data, 8)

	// Step 4: Interpret as float64 array
	nElements := rows * cols
	grid := make([]float64, nElements)
	for i := 0; i < nElements; i++ {
		bits := binary.LittleEndian.Uint64(unshuffled[i*8 : (i+1)*8])
		v := math.Float64frombits(bits)
		if math.IsNaN(v) || math.IsInf(v, 0) {
			v = 0
		}
		grid[i] = v
	}

	return &RadarGrid{
		Rows: rows, Cols: cols, Data: grid,
		MinLat: llLat, MaxLat: urLat, MinLon: llLon, MaxLon: urLon,
	}, nil
}

func getAttrInt(attrs api.AttributeMap, key string) int {
	v, ok := attrs.Get(key)
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int32:
		return int(val)
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

func getAttrFloat(attrs api.AttributeMap, key string) float64 {
	v, ok := attrs.Get(key)
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	default:
		return 0
	}
}

// findAndDecompressChunk scans an HDF5 file for a zlib-compressed block
// that decompresses to exactly expectedSize bytes.
func findAndDecompressChunk(raw []byte, expectedSize int) ([]byte, error) {
	// Scan for zlib headers: 0x78 followed by 0x01, 0x5E, 0x9C, 0xDA, or 0xBB
	validSecond := map[byte]bool{0x01: true, 0x5E: true, 0x9C: true, 0xDA: true, 0xBB: true}

	for i := 0; i < len(raw)-2; i++ {
		if raw[i] != 0x78 || !validSecond[raw[i+1]] {
			continue
		}

		// Try decompressing from this offset
		reader, err := zlib.NewReader(bytes.NewReader(raw[i:]))
		if err != nil {
			continue
		}
		decompressed, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			continue
		}

		if len(decompressed) == expectedSize {
			return decompressed, nil
		}
	}

	return nil, fmt.Errorf("could not find compressed data chunk (expected %d bytes decompressed)", expectedSize)
}

// unshuffleBytes reverses the HDF5 byte shuffle filter.
// The shuffle groups all first bytes of elements together, then second bytes, etc.
func unshuffleBytes(data []byte, elementSize int) []byte {
	nElements := len(data) / elementSize
	result := make([]byte, len(data))
	for i := 0; i < elementSize; i++ {
		for j := 0; j < nElements; j++ {
			result[j*elementSize+i] = data[i*nElements+j]
		}
	}
	return result
}

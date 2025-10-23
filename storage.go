package lib

import (
	"encoding/json"
	"io"

	"github.com/taubyte/go-sdk/event"
	http "github.com/taubyte/go-sdk/http/event"
	"github.com/taubyte/go-sdk/storage"
)

func failed(h http.Event, err error, code int) uint32 {
	h.Write([]byte(err.Error()))
	h.Return(code)
	return 1
}

func setCORSHeaders(h http.Event) {
	h.Headers().Set("Access-Control-Allow-Origin", "*")
	h.Headers().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	h.Headers().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

type UploadReq struct {
	Filename string `json:"filename"`
	Data     string `json:"data"`
}

// POST /api/upload
//export upload
func upload(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Open/Create the storage
	sto, err := storage.New("/storage")
	if err != nil {
		return failed(h, err, 500)
	}

	// Read the request body
	reqDec := json.NewDecoder(h.Body())
	defer h.Body().Close()

	var req UploadReq
	err = reqDec.Decode(&req)
	if err != nil {
		return failed(h, err, 500)
	}

	// Select file/object
	file := sto.File(req.Filename)

	// Write data to the file
	_, err = file.Add([]byte(req.Data), true)
	if err != nil {
		return failed(h, err, 500)
	}

	h.Write([]byte("File uploaded successfully"))
	h.Return(200)
	return 0
}

// GET /api/download?filename={filename}
//export download
func download(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Read the filename from the query string
	filename, err := h.Query().Get("filename")
	if err != nil {
		return failed(h, err, 400)
	}

	// Open/Create the storage
	sto, err := storage.New("/storage")
	if err != nil {
		return failed(h, err, 500)
	}

	// Select file/object
	file := sto.File(filename)

	// Get a io.ReadCloser
	reader, err := file.GetFile()
	if err != nil {
		return failed(h, err, 500)
	}
	defer reader.Close()

	// Read from file and write to response
	_, err = io.Copy(h, reader)
	if err != nil {
		return failed(h, err, 500)
	}

	h.Return(200)
	return 0
}

// GET /api/list
//export listFiles
func listFiles(e event.Event) uint32 {
	h, err := e.HTTP()
	if err != nil {
		return 1
	}

	setCORSHeaders(h)

	// Open/Create the storage
	sto, err := storage.New("/storage")
	if err != nil {
		return failed(h, err, 500)
	}

	// List all files
	files, err := sto.ListFiles()
	if err != nil {
		// Return empty array on error
		emptyArray := []string{}
		filesJson, _ := json.Marshal(emptyArray)
		h.Write(filesJson)
		h.Return(200)
		return 0
	}

	// Convert files to string array if needed
	var fileNames []string
	if len(files) > 0 {
		// Check if files are already strings or objects
		for _, file := range files {
			if fileName, ok := file.(string); ok {
				fileNames = append(fileNames, fileName)
			} else {
				// If it's an object, try to extract the name field
				if fileObj, ok := file.(map[string]interface{}); ok {
					if name, exists := fileObj["name"]; exists {
						if nameStr, ok := name.(string); ok {
							fileNames = append(fileNames, nameStr)
						}
					}
				}
			}
		}
	}

	// Return list of filenames
	filesJson, err := json.Marshal(fileNames)
	if err != nil {
		emptyArray := []string{}
		filesJson, _ = json.Marshal(emptyArray)
	}

	h.Write(filesJson)
	h.Return(200)
	return 0
}

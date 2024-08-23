package main

import (
	"io/fs"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"

	_ "github.com/tliron/commonlog/simple"
)

const lsName = "glsl_lsp_go"

var version string = "0.0.1"
var handler protocol.Handler

var pathMutex sync.Mutex
var currentDir protocol.URI

func main() {
	path := "apth_lsp.log"
	commonlog.Configure(2, &path)
	commonlog.NewInfoMessage(0, "Startup").Send()

	handler = protocol.Handler{
		Initialize:             initialize,
		Shutdown:               shutdown,
		TextDocumentDidOpen:    documentOpen,
		TextDocumentDidSave:    documentSave,
		TextDocumentDidChange:  documentChange,
		TextDocumentCompletion: completion,
	}

	server := server.NewServer(&handler, lsName, true)

	server.RunStdio()
}

func completion(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
	var completions []protocol.CompletionItem

	filepath.WalkDir(
		currentDir,
		func(path string, d fs.DirEntry, err error) error {
			append(completions, protocol.CompletionItem{
				Label:               "",
				Detail:              new(string),
				InsertTextFormat:    &0,
				InsertTextMode:      &0,
				TextEdit:            nil,
				AdditionalTextEdits: []protocol.TextEdit{},
				CommitCharacters:    []string{},
				Command:             &protocol.Command{},
				Data:                d,
			})
			return nil
		},
	)

	return completions, nil
}

func initialize(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
	commonlog.NewInfoMessage(0, "Initializing server...").Send()

	capabilities := handler.CreateServerCapabilities()

	capabilities.TextDocumentSync = protocol.TextDocumentSyncKindFull
	capabilities.CompletionProvider = &protocol.CompletionOptions{}

	return protocol.InitializeResult{
		Capabilities: capabilities,
		ServerInfo: &protocol.InitializeResultServerInfo{
			Name:    lsName,
			Version: &version,
		},
	}, nil
}

func shutdown(context *glsp.Context) error {
	return nil
}

func documentOpen(context *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
	pathMutex.Lock()
	defer pathMutex.Unlock()
	currentDir = path.Dir(params.TextDocument.URI)

	return nil
}

func documentChange(context *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
	pathMutex.Lock()
	defer pathMutex.Unlock()
	currentDir = path.Dir(params.TextDocument.URI)
	return nil
}

// returns line number and offset of last newline before psoition
func convertToLine(code string, position int) (int, int) {
	until := string([]rune(code)[0:position])
	lineCount := strings.Count(until, "\n")
	offset := strings.LastIndex(until, "\n")
	return lineCount, offset
}

func documentSave(context *glsp.Context, params *protocol.DidSaveTextDocumentParams) error {
	pathMutex.Lock()
	defer pathMutex.Unlock()
	currentDir = path.Dir(params.TextDocument.URI)
	return nil
}

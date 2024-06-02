// Copyright (C) 2024 Ronsor Labs. Licensed under the MIT license.

package rwkvtkn

import (
	"bufio"
	"bytes"
	_ "embed"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

var (
	ErrMalformedVocabulary = errors.New("malformed tokenizer vocabulary")
	ErrUnknownToken        = errors.New("unknown token ID")
	ErrCannotTokenize      = errors.New("cannot tokenize data")
)

type trieNode struct {
	children [256]*trieNode
	value    int
}

func (t *trieNode) Insert(key []byte, value int) {
	node := t
	for _, c := range key {
		ci := c
		if node.children[ci] == nil {
			node.children[ci] = &trieNode{value: -1}
		}

		node = node.children[ci]
	}
	node.value = value
}

func (t *trieNode) InsertString(key string, value int) {
	t.Insert([]byte(key), value)
}

func (t *trieNode) FindLongest(data []byte, index int) (endIndex, value int) {
	node := t
	endIndex, value = 0, -1
	for node.children[data[index]] != nil {
		node = node.children[data[index]]
		index += 1

		if node.value != -1 {
			endIndex = index
			value = node.value
		}

		if index == len(data) {
			break
		}
	}
	return
}

// Tokenizer is a trie-based RWKV tokenizer.
type Tokenizer struct {
	trie *trieNode
	t2i  map[string]int
	i2t  map[int]string
}

// NewTokenizer creates a new Tokenizer with an empty vocabulary.
func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		trie: &trieNode{value: -1},
		t2i:  make(map[string]int),
		i2t:  make(map[int]string),
	}
}

var singleUnescapeFixer = strings.NewReplacer("\"", "\\\"", "\\'", "'")
var unicodeUnescapeFixer = strings.NewReplacer("\\x", "\\u00")

// NewTokenizer creates a new Tokenizer whose vocabulary is read from
// the supplied io.Reader.
func NewTokenizerFromReader(r io.Reader) (*Tokenizer, error) {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}

	t := NewTokenizer()
	for {
		line, err := br.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		line = strings.Trim(line, " \v\t\r\n")
		if line == "" || line[0] == '#' {
			continue
		}

		sl, sr := strings.IndexByte(line, ' '), strings.LastIndexByte(line, ' ')
		if sl == sr || sr == len(line)-1 {
			return nil, ErrMalformedVocabulary
		}

		id, err := strconv.Atoi(line[:sl])
		if err != nil {
			return nil, err
		}

		tokLit := strings.Trim(line[sl:sr], " \v\t\r\n")
		tokIsByt := len(tokLit) > 0 && tokLit[0] == 'b'
		if tokIsByt {
			tokLit = tokLit[1:]
		}

		if len(tokLit) < 2 || tokLit[0] != tokLit[len(tokLit)-1] {
			return nil, ErrMalformedVocabulary
		}

		switch tokLit[0] {
		case '"':
		case '\'':
			tokLit = "\"" + singleUnescapeFixer.Replace(tokLit[1:len(tokLit)-1]) + "\""
		default:
			return nil, ErrMalformedVocabulary
		}

		if !tokIsByt {
			tokLit = unicodeUnescapeFixer.Replace(tokLit)
		}

		tokStr, err := strconv.Unquote(tokLit)
		if err != nil {
			return nil, err
		}

		tokLen, err := strconv.Atoi(line[sr+1:])
		if err != nil {
			return nil, err
		} else if tokLen != len(tokStr) {
			return nil, ErrMalformedVocabulary
		}

		t.AddTokenString(tokStr, id)
	}
	return t, nil
}

// NewTokenizer creates a new Tokenizer whose vocabulary is read from
// the specified file.
func NewTokenizerFromFile(path string) (*Tokenizer, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return NewTokenizerFromReader(f)
}

//go:embed rwkv_vocab_v20230424.txt
var rwkvVocab20230424 []byte

// NewWorldTokenizer creates a new Tokenizer with the default RWKV World
// vocabulary (rwkv_vocab_20230424).
func NewWorldTokenizer() *Tokenizer {
	t, err := NewTokenizerFromReader(bytes.NewReader(rwkvVocab20230424))
	if err != nil {
		panic(err.Error())
	}
	return t
}

// AddToken adds a token, represented as a byte slice, to the Tokenizer's
// vocabulary.
func (t *Tokenizer) AddToken(token []byte, id int) {
	t.trie.Insert(token, id)

	t.t2i[string(token)] = id
	t.i2t[id] = string(token)
}

// AddTokenString adds a token, represented as a string, to the Tokenizer's
// vocabulary.
func (t *Tokenizer) AddTokenString(token string, id int) {
	t.trie.InsertString(token, id)

	t.t2i[token] = id
	t.i2t[id] = token
}

// Encode encodes the given byte slice into an int slice of tokens.
func (t *Tokenizer) Encode(data []byte) (tokens []int, err error) {
	n := 0
	tokens = make([]int, 0, 32)
	for n < len(data) {
		n2, id := t.trie.FindLongest(data, n)
		if n2 == n || id == -1 {
			return tokens, ErrCannotTokenize
		}
		tokens = append(tokens, id)
		n = n2
	}
	return
}

// EncodeString encodes the given string into an int slice of tokens.
func (t *Tokenizer) EncodeString(text string) (tokens []int, err error) {
	return t.Encode([]byte(text))
}

// Decode decodes an int slice of tokens to a byte slice.
func (t *Tokenizer) Decode(tokens []int) (data []byte, err error) {
	var b bytes.Buffer
	for _, v := range tokens {
		if tokStr, ok := t.i2t[v]; ok {
			b.WriteString(tokStr)
		} else {
			err = ErrUnknownToken
		}
	}
	data = b.Bytes()
	return
}

// Decode decodes an int slice of tokens to a string.
func (t *Tokenizer) DecodeToString(tokens []int) (text string, err error) {
	var b strings.Builder
	for _, v := range tokens {
		if tokStr, ok := t.i2t[v]; ok {
			b.WriteString(tokStr)
		} else {
			err = ErrUnknownToken
		}
	}
	text = b.String()
	return
}

// TokenToID returns the ID of the specified token.
func (t *Tokenizer) TokenToID(token string) (int, error) {
	if id, ok := t.t2i[token]; ok {
		return id, nil
	} else {
		return -1, ErrCannotTokenize
	}
}

// IDToToken returns the token for the given ID.
func (t *Tokenizer) IDToToken(id int) (string, error) {
	if token, ok := t.i2t[id]; ok {
		return token, nil
	} else {
		return "", ErrCannotTokenize
	}
}

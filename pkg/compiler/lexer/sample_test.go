package lexer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"io"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

// convertShiftJISToUTF8 converts Shift-JIS encoded data to UTF-8.
func convertShiftJISToUTF8(data []byte) (string, error) {
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(utf8Data), nil
}

// TestSampleFileROBOT tests the Lexer with the sample file samples/robot/ROBOT.TFY.
// This test validates that the Lexer can correctly tokenize a real FILLY script.
// Validates Requirements 2.1-2.14: Complete lexical analysis of FILLY scripts.
func TestSampleFileROBOT(t *testing.T) {
	// Find the sample file relative to the workspace root
	samplePath := filepath.Join("..", "..", "..", "samples", "robot", "ROBOT.TFY")

	// Read the sample file
	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	// Convert from Shift-JIS to UTF-8
	content, err := convertShiftJISToUTF8(data)
	if err != nil {
		t.Fatalf("Failed to convert Shift-JIS to UTF-8: %v", err)
	}

	// Create lexer and tokenize
	l := New(content)
	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	// Verify that we got tokens
	if len(tokens) == 0 {
		t.Fatal("Expected tokens, got none")
	}

	// Count token types
	tokenCounts := make(map[TokenType]int)
	illegalTokens := []Token{}

	for _, tok := range tokens {
		tokenCounts[tok.Type]++
		if tok.Type == TOKEN_ILLEGAL {
			illegalTokens = append(illegalTokens, tok)
		}
	}

	// Log token statistics
	t.Logf("Total tokens: %d", len(tokens))
	t.Logf("Token type counts:")
	for tokType, count := range tokenCounts {
		t.Logf("  %s: %d", tokType.String(), count)
	}

	// Check for ILLEGAL tokens
	// Note: # directives (#info, #include) are expected to produce ILLEGAL tokens
	// because they are preprocessor directives not handled by the lexer
	expectedIllegalCount := 0
	for _, tok := range illegalTokens {
		if tok.Literal == "#" {
			expectedIllegalCount++
		}
	}

	unexpectedIllegalTokens := []Token{}
	for _, tok := range illegalTokens {
		if tok.Literal != "#" {
			unexpectedIllegalTokens = append(unexpectedIllegalTokens, tok)
		}
	}

	if len(unexpectedIllegalTokens) > 0 {
		t.Errorf("Found %d unexpected ILLEGAL tokens:", len(unexpectedIllegalTokens))
		for _, tok := range unexpectedIllegalTokens {
			t.Errorf("  Line %d, Column %d: %q", tok.Line, tok.Column, tok.Literal)
		}
	}

	t.Logf("Expected ILLEGAL tokens (# directives): %d", expectedIllegalCount)

	// Verify key tokens are present
	verifyKeyTokensPresent(t, tokens)
}

// verifyKeyTokensPresent checks that key tokens from the sample file are correctly identified.
func verifyKeyTokensPresent(t *testing.T, tokens []Token) {
	t.Helper()

	// Expected keywords that should be found in the sample file
	expectedKeywords := map[TokenType]bool{
		TOKEN_INT_TYPE: false, // int
		TOKEN_FOR:      false, // for
		TOKEN_IF:       false, // if
		TOKEN_MES:      false, // mes
		TOKEN_STEP:     false, // step
		TOKEN_END_STEP: false, // end_step
		TOKEN_DEL_ME:   false, // del_me
		TOKEN_DEL_ALL:  false, // del_all
	}

	// Expected identifiers that should be found
	expectedIdents := map[string]bool{
		"main":      false,
		"start":     false,
		"LoadPic":   false,
		"CreatePic": false,
		"MovePic":   false,
		"OpenWin":   false,
		"CloseWin":  false,
		"TIME":      false,
		"MIDI_TIME": false,
		"MIDI_END":  false,
	}

	// Check for expected tokens
	for _, tok := range tokens {
		// Check keywords
		if _, ok := expectedKeywords[tok.Type]; ok {
			expectedKeywords[tok.Type] = true
		}

		// Check identifiers
		if tok.Type == TOKEN_IDENT {
			if _, ok := expectedIdents[tok.Literal]; ok {
				expectedIdents[tok.Literal] = true
			}
		}
	}

	// Report missing keywords
	for tokType, found := range expectedKeywords {
		if !found {
			t.Errorf("Expected keyword %s not found in tokens", tokType.String())
		}
	}

	// Report missing identifiers
	for ident, found := range expectedIdents {
		if !found {
			t.Errorf("Expected identifier %q not found in tokens", ident)
		}
	}
}

// TestSampleFileTokenSequence tests specific token sequences from the sample file.
func TestSampleFileTokenSequence(t *testing.T) {
	// Test a specific code snippet from the sample file
	// This is the main function declaration pattern
	input := `main(){
  CapTitle("");
  WinW=WinInfo(0); WinH=WinInfo(1);
}`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "main"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "CapTitle"},
		{TOKEN_LPAREN, "("},
		{TOKEN_STRING, ""},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_IDENT, "WinW"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_IDENT, "WinInfo"},
		{TOKEN_LPAREN, "("},
		{TOKEN_INT, "0"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_IDENT, "WinH"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_IDENT, "WinInfo"},
		{TOKEN_LPAREN, "("},
		{TOKEN_INT, "1"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestMesBlockTokenization tests tokenization of mes() blocks.
// Validates Requirement 2.2: Keywords are identified case-insensitively.
func TestMesBlockTokenization(t *testing.T) {
	input := `mes(TIME){step(20){,start();end_step;del_me;}}`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_MES, "mes"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "TIME"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_STEP, "step"},
		{TOKEN_LPAREN, "("},
		{TOKEN_INT, "20"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "start"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_END_STEP, "end_step"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_DEL_ME, "del_me"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestStepBlockWithCommas tests tokenization of step blocks with commas as wait commands.
// Validates Requirement 2.8: Delimiters including comma are correctly tokenized.
func TestStepBlockWithCommas(t *testing.T) {
	input := `step{BIRD();,,,, OPENON();,,,, end_step; del_me;}`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_STEP, "step"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "BIRD"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "OPENON"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_COMMA, ","},
		{TOKEN_END_STEP, "end_step"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_DEL_ME, "del_me"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestHexColorLiterals tests tokenization of hex color values like 0xffffff.
// Validates Requirement 2.4: Hexadecimal integers with 0x prefix create INT tokens.
func TestHexColorLiterals(t *testing.T) {
	input := `OpenWin(BirdPic[0],0,0,WinW,WinH,WinX,WinY,0xffffff)`

	l := New(input)

	// Find the hex color token
	var hexToken Token
	for {
		tok := l.NextToken()
		if tok.Type == TOKEN_EOF {
			break
		}
		if tok.Type == TOKEN_INT && strings.HasPrefix(tok.Literal, "0x") {
			hexToken = tok
			break
		}
	}

	if hexToken.Type != TOKEN_INT {
		t.Errorf("Expected hex color token, got type=%v", hexToken.Type)
	}
	if hexToken.Literal != "0xffffff" {
		t.Errorf("Expected literal='0xffffff', got=%q", hexToken.Literal)
	}
}

// TestVariableDeclarations tests tokenization of variable declarations.
// Validates Requirements 2.2, 2.3: Keywords and identifiers are correctly identified.
func TestVariableDeclarations(t *testing.T) {
	input := `int LPic[],BasePic,FieldPic,BirdPic[],OPPic[],Dummy;`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_INT_TYPE, "int"},
		{TOKEN_IDENT, "LPic"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "BasePic"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "FieldPic"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "BirdPic"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "OPPic"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "Dummy"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestForLoopTokenization tests tokenization of for loops.
// Validates Requirements 2.2, 2.7, 2.8: Keywords, operators, and delimiters.
func TestForLoopTokenization(t *testing.T) {
	input := `for(i=0;i<=1;i=i+1){
    LPic[i]=LoadPic(StrPrint("ROBOT%03d.BMP",i));
  }`

	l := New(input)

	// Check first few tokens
	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_FOR, "for"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "i"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "0"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_IDENT, "i"},
		{TOKEN_LTE, "<="},
		{TOKEN_INT, "1"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_IDENT, "i"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_IDENT, "i"},
		{TOKEN_PLUS, "+"},
		{TOKEN_INT, "1"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestFunctionWithDefaultParams tests tokenization of functions with default parameters.
// Validates Requirements 2.2, 2.3, 2.7: Keywords, identifiers, and operators.
func TestFunctionWithDefaultParams(t *testing.T) {
	input := `OP_walk(c,p[],x,y,w,h,l=10){}`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "OP_walk"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "c"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "p"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "x"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "y"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "w"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "h"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "l"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestMultiLineCommentFromSample tests multi-line comments as used in the sample file.
// Validates Requirement 2.10: Multi-line comments are skipped.
func TestMultiLineCommentFromSample(t *testing.T) {
	input := `/* 走り回るロボット */

/* 作品情報 */
int x;`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_INT_TYPE, "int"},
		{TOKEN_IDENT, "x"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v (literal=%q)", i, exp.tokenType, tok.Type, tok.Literal)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestSingleLineCommentFromSample tests single-line comments as used in the sample file.
// Validates Requirement 2.9: Single-line comments are skipped.
// Validates Requirement 2.17: #include directive extracts the filename.
func TestSingleLineCommentFromSample(t *testing.T) {
	input := `#include "MINIF_.TFY"	// Copyright (C) 作者
int x;`

	l := New(input)

	// #include should be parsed as INCLUDE token with filename
	tok := l.NextToken()
	if tok.Type != TOKEN_INCLUDE {
		t.Errorf("Expected INCLUDE, got type=%v literal=%q", tok.Type, tok.Literal)
	}
	// The literal should be the filename (quotes removed)
	if tok.Literal != "MINIF_.TFY" {
		t.Errorf("Expected literal 'MINIF_.TFY', got %q", tok.Literal)
	}

	// After the comment, we should get int
	tok = l.NextToken()
	if tok.Type != TOKEN_INT_TYPE {
		t.Errorf("Expected INT_TYPE, got type=%v literal=%q", tok.Type, tok.Literal)
	}
}

// TestNestedMesBlocks tests tokenization of nested mes blocks.
func TestNestedMesBlocks(t *testing.T) {
	input := `mes(MIDI_TIME){step{
    mes(MIDI_TIME){step(8){
      OPENING();,
      end_step; del_me;
    }}end_step; del_me;
  }}`

	l := New(input)
	tokens, _ := l.Tokenize()

	// Count mes keywords
	mesCount := 0
	stepCount := 0
	endStepCount := 0
	delMeCount := 0

	for _, tok := range tokens {
		switch tok.Type {
		case TOKEN_MES:
			mesCount++
		case TOKEN_STEP:
			stepCount++
		case TOKEN_END_STEP:
			endStepCount++
		case TOKEN_DEL_ME:
			delMeCount++
		}
	}

	if mesCount != 2 {
		t.Errorf("Expected 2 mes keywords, got %d", mesCount)
	}
	if stepCount != 2 {
		t.Errorf("Expected 2 step keywords, got %d", stepCount)
	}
	if endStepCount != 2 {
		t.Errorf("Expected 2 end_step keywords, got %d", endStepCount)
	}
	if delMeCount != 2 {
		t.Errorf("Expected 2 del_me keywords, got %d", delMeCount)
	}
}

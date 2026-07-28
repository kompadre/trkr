package fm

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

const (
	VoiceSize       = 128
	VoicesPerBank   = 32
	BankDataSize    = VoiceSize * VoicesPerBank // 4096
	SysexHeaderSize = 6
	SysexFooterSize = 1
	FullSysexSize   = SysexHeaderSize + BankDataSize + 1 + SysexFooterSize // 4104
)

var SysexHeader = []byte{0xF0, 0x43, 0x00, 0x09, 0x20, 0x00}
var SysexFooter = []byte{0xF7}

type Voice [VoiceSize]byte

func (v *Voice) Name() string {
	// DX7 names are the last 10 bytes
	nameBytes := v[118:128]
	var sb strings.Builder
	for _, b := range nameBytes {
		if b >= 32 && b <= 126 {
			sb.WriteByte(b)
		} else {
			sb.WriteByte(' ')
		}
	}
	return strings.TrimSpace(sb.String())
}

func (v *Voice) IsEmpty() bool {
	name := v.Name()
	if name == "" || strings.HasPrefix(strings.ToUpper(name), "INIT ") || strings.HasPrefix(strings.ToUpper(name), "INITVOICE") {
		return true
	}
	// Check if data is mostly zeros (excluding some header/footer bytes if applicable, 
	// but DX7 voices usually have non-zero data even if init)
	// For now, name check is usually enough for DX7 banks.
	return false
}

func (v *Voice) Clear() {
	for i := range v {
		v[i] = 0
	}
	// Standard DX7 Init Voice name is "INIT VOICE"
	copy(v[118:128], "INIT VOICE")
}

type Bank struct {
	Voices [VoicesPerBank]Voice
}

func CalculateChecksum(data []byte) byte {
	var sum int
	for _, b := range data {
		sum += int(b)
	}
	return byte((^sum + 1) & 0x7F)
}

func ParseSysex(data []byte) (*Bank, error) {
	if len(data) < FullSysexSize {
		return nil, fmt.Errorf("sysex data too short: %d bytes", len(data))
	}

	// Verify header
	for i := 0; i < len(SysexHeader); i++ {
		if data[i] != SysexHeader[i] {
			return nil, errors.New("invalid DX7 bulk dump header")
		}
	}

	bank := &Bank{}
	offset := SysexHeaderSize
	for i := 0; i < VoicesPerBank; i++ {
		copy(bank.Voices[i][:], data[offset:offset+VoiceSize])
		offset += VoiceSize
	}

	return bank, nil
}

func (b *Bank) ToSysex() []byte {
	data := make([]byte, FullSysexSize)
	copy(data[0:SysexHeaderSize], SysexHeader)
	
	offset := SysexHeaderSize
	for i := 0; i < VoicesPerBank; i++ {
		copy(data[offset:offset+VoiceSize], b.Voices[i][:])
		offset += VoiceSize
	}

	checksum := CalculateChecksum(data[SysexHeaderSize : SysexHeaderSize+BankDataSize])
	data[FullSysexSize-2] = checksum
	data[FullSysexSize-1] = SysexFooter[0]

	return data
}

func SanitizeFilename(name string) string {
	// Strip extension if present
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}

	var sb strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			sb.WriteRune(r)
		} else if unicode.IsSpace(r) {
			sb.WriteRune('_')
		}
	}
	result := sb.String()
	if result == "" {
		return "untitled"
	}
	return result
}

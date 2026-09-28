package cache

import (
	"strconv"
	"strings"
	"time"
	"uuid"
)

const partSeparator = "\x1f"

type KeyPart interface {
	writeTo(b *strings.Builder)
}

func Str(s string) KeyPart      { return strPart(s) }
func Int(i int) KeyPart         { return intPart(i) }
func Bool(v bool) KeyPart       { return boolPart(v) }
func UUID(id uuid.UUID) KeyPart { return uuidPart(id) }
func Time(t time.Time) KeyPart  { return timePart(t) }

func OptStr(p *string) KeyPart {
	if p == nil {
		return strPart("")
	}
	return strPart(*p)
}

func OptInt(p *int) KeyPart {
	if p == nil {
		return intPart(0)
	}
	return intPart(*p)
}

func OptBool(p *bool) KeyPart {
	if p == nil {
		return boolPart(false)
	}
	return boolPart(*p)
}

func OptUUID(p *uuid.UUID) KeyPart {
	if p == nil {
		return strPart("")
	}
	return uuidPart(*p)
}

func OptTime(p *time.Time) KeyPart {
	if p == nil {
		return strPart("")
	}
	return timePart(*p)
}

type strPart string

func (s strPart) writeTo(b *strings.Builder) {
	b.WriteString(partSeparator)
	b.WriteString(string(s))
}

type intPart int

func (i intPart) writeTo(b *strings.Builder) {
	b.WriteString(partSeparator)
	b.WriteString(strconv.Itoa(int(i)))
}

type boolPart bool

func (v boolPart) writeTo(b *strings.Builder) {
	b.WriteString(partSeparator)
	if bool(v) {
		b.WriteByte('1')
	} else {
		b.WriteByte('0')
	}
}

type uuidPart uuid.UUID

func (id uuidPart) writeTo(b *strings.Builder) {
	b.WriteString(partSeparator)
	b.WriteString(uuid.UUID(id).String())
}

type timePart time.Time

func (t timePart) writeTo(b *strings.Builder) {
	b.WriteString(partSeparator)
	b.WriteString(time.Time(t).UTC().Format(time.RFC3339))
}

type KeyBuilder interface {
	BuildItemKey(id string) string

	BuildListKey(parts ...KeyPart) string

	GetPrefix() string
}

type keyBuilder struct {
	prefix string
}

func NewKeyBuilder(prefix string) *keyBuilder {
	return &keyBuilder{prefix: prefix}
}

func (kb *keyBuilder) GetPrefix() string {
	return kb.prefix
}

func (kb *keyBuilder) BuildItemKey(id string) string {
	var b strings.Builder
	b.Grow(len(kb.prefix) + len(id) + 8)
	b.WriteString(kb.prefix)
	b.WriteString(":item")
	b.WriteString(partSeparator)
	b.WriteString(id)
	return b.String()
}

func (kb *keyBuilder) BuildListKey(parts ...KeyPart) string {
	var b strings.Builder
	b.WriteString(kb.prefix)
	b.WriteString(":list")
	for _, p := range parts {
		p.writeTo(&b)
	}
	return b.String()
}

package main // TODO: replace this file with "embed", when it will be released: https://github.com/golang/go/issues/41191
const templateMethods = `// +build !manifestcodegen 
// Code generated by "menifestcodegen". DO NOT EDIT.
// To reproduce: go run github.com/9elements/converged-security-suite/pkg/intel/metadata/manifest/internal/manifestcodegen/cmd/manifestcodegen {{ .Package.Path }}

package {{ .Package.Name }}

import (
{{- if not .EnableTracing }}
	"encoding/binary"
{{- else }}
	binary "github.com/9elements/converged-security-suite/pkg/intel/metadata/manifest/internal/tracedbinary"
{{- end }}
	"fmt"
	"io"
	"strings"

	"github.com/9elements/converged-security-suite/pkg/intel/metadata/manifest/internal/pretty"
{{- if ne .Package.Name "manifest" }}
	"github.com/9elements/converged-security-suite/pkg/intel/metadata/manifest"
{{- end }}
)

var (
	// Just to avoid errors in "import" above in case if it wasn't used below
	_ = binary.LittleEndian
	_ = (fmt.Stringer)(nil)
	_ = (io.Reader)(nil)
	_ = pretty.Header
	_ = strings.Join
{{- if ne .Package.Name "manifest" }}
	_ = manifest.StructInfo{}
{{- end }}
)
{{- $manifestRootPath := ternary (ne .Package.Name "manifest") "manifest." "" }}
{{- $enableTracing := .EnableTracing }}

{{- range $index, $struct := .Structs }}

// New{{ $struct.Name }} returns a new instance of {{ $struct.Name }} with
// all default values set.
func New{{ $struct.Name }}() *{{ $struct.Name }} {
	s := &{{ $struct.Name }}{}
 {{- if ne $struct.ElementStructID "" }}
	copy(s.StructInfo.ID[:], []byte(StructureID{{ $struct.Name }}))
	s.StructInfo.Version = {{ $struct.ElementStructVersion }}
 {{- end }}
 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}
  {{- if and (not $field.IsSlice) (not $field.IsPointer) (or (eq $fieldType "element") (eq $fieldType "subStruct")) }}
	// Recursively initializing a child structure:
	s.{{ $field.Name }} = *{{ $field.AccessPrefix }}New{{ $field.ItemTypeName }}()
  {{- end }}
  {{- $defaultValue := ternary (ne $field.RequiredValue "") $field.RequiredValue $field.DefaultValue }}
  {{- $defaultValueSource := ternary (ne $field.RequiredValue "") "required" "default" }}
  {{- if ne $defaultValue "" }}
   {{- if eq $fieldType "arrayStatic" }}
    {{- if ne $defaultValue "0" }}
	// Set through tag "{{ $defaultValueSource }}":
	for idx := range s.{{ $field.Name }} {
		s.{{ $field.Name }}[idx] = {{ $defaultValue }}
	}
    {{- end }}
   {{- else }}
	// Set through tag "{{ $defaultValueSource }}":
	s.{{ $field.Name }} = {{ $defaultValue }}
   {{- end }}
  {{- end }}
 {{- end }}
	s.Rehash()
	return s
}

// Validate (recursively) checks the structure if there are any unexpected
// values. It returns an error if so.
func (s *{{ $struct.Name }}) Validate() error {
 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}
  {{- if and (not $field.IsSlice) (not $field.IsPointer) (or (eq $fieldType "element") (eq $fieldType "subStruct")) }}
	// Recursively validating a child structure:
	if err := s.{{ $field.Name }}.Validate(); err != nil {
		return fmt.Errorf("error on field '{{ $field.Name }}': %w", err)
	}
  {{- end }}
  {{- if ne $field.RequiredValue "" }}
	// See tag "require" 
   {{- if eq $fieldType "arrayStatic" }}
	for idx := range s.{{ $field.Name }} {
		if s.{{ $field.Name }}[idx] != {{ $field.RequiredValue }} {
			return fmt.Errorf("'{{ $field.Name }}[%d]' is expected to be {{ $field.RequiredValue }}, but it is %v", idx, s.{{ $field.Name }}[idx])
		}
	}
   {{- else }}
	if s.{{ $field.Name }} != {{ $field.RequiredValue }} {
		return fmt.Errorf("field '{{ $field.Name }}' expects value '{{ $field.RequiredValue }}', but has %v", s.{{ $field.Name }})
	}
   {{- end }}
  {{- end }}	
  {{- if ne $field.RehashValue "" }}
	// See tag "rehashValue" 
	{
		expectedValue := {{ $field.Type }}(s.{{ $field.RehashValue }})
		if s.{{ $field.Name }} != expectedValue {
			return fmt.Errorf("field '{{ $field.Name }}' expects write-value '%v', but has %v", expectedValue, s.{{ $field.Name }})
		}
	}
  {{- end }}
 {{- end }}

	return nil
}

{{- if ne $struct.ElementStructID "" }}
// StructureID{{ $struct.Name }} is the StructureID (in terms of
// the document #575623) of element '{{ $struct.Name }}'. 
const StructureID{{ $struct.Name }} = "{{ $struct.ElementStructID }}"

// GetStructInfo returns current value of StructInfo of the structure.
//
// StructInfo is a set of standard fields with presented in any element
// ("element" in terms of document #575623).
func (s *{{ $struct.Name }}) GetStructInfo() {{ $manifestRootPath }}StructInfo {
	return s.StructInfo
}

// SetStructInfo sets new value of StructInfo to the structure.
//
// StructInfo is a set of standard fields with presented in any element
// ("element" in terms of document #575623).
func (s *{{ $struct.Name }}) SetStructInfo(newStructInfo {{ $manifestRootPath }}StructInfo) {
	s.StructInfo = newStructInfo
}

{{- end }}

{{- if $struct.IsElementsContainer }}

// fieldIndexByStructID returns the position index within
// structure {{ $struct.Name }} of the field by its StructureID
// (see document #575623, an example of StructureID value is "__KEYM__"). 
func (_ {{ $struct.Name }}) fieldIndexByStructID(structID string) int {
	switch structID {
 {{- range $index, $field := $struct.Fields }}
	case StructureID{{ $field.Struct.Name }}:
		return {{ $index }}
 {{- end }}	
	}

	return -1
}

// fieldNameByIndex returns the name of the field by its position number
// within structure {{ $struct.Name }}. 
func (_ {{ $struct.Name }}) fieldNameByIndex(fieldIndex int) string {
	switch fieldIndex {
 {{- range $index, $field := $struct.Fields }}
	case {{ $index }}:
		return "{{ $field.Name }}"
 {{- end }}	
	}

	return fmt.Sprintf("invalidFieldIndex_%d", fieldIndex)
}

// ReadFrom reads the {{ $struct.Name }} from 'r' in format defined in the document #575623.
func (s *{{ $struct.Name }}) ReadFrom(r io.Reader) (int64, error) {
	var missingFieldsByIndices = [{{ len $struct.Fields }}]bool{
 {{- range $index, $field := $struct.Fields }}
  {{- if and (not $field.IsSlice) (not $field.IsPointer) }}
		{{ $index }}: true,
  {{- end }}
 {{- end }}
	}
	var totalN int64
	previousFieldIndex := int(-1)
	for {
		var structInfo {{ $manifestRootPath }}StructInfo
		err := binary.Read(r, binary.LittleEndian, &structInfo)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return totalN, nil
		}
		if err != nil {
			return totalN, fmt.Errorf("unable to read structure info at %d: %w", totalN, err)
		}
    	{{- if $enableTracing }}
		fmt.Printf("%s header parsed, TotalN is %d -> %d\n", structInfo.ID.String(), totalN, totalN + int64(binary.Size(structInfo))){{- end}}
		totalN += int64(binary.Size(structInfo))

		structID := structInfo.ID.String()
		fieldIndex := s.fieldIndexByStructID(structID)
		if fieldIndex < 0 {
			// TODO: report error "unknown structure ID: '"+structID+"'"
			continue
		}
		if {{ $manifestRootPath }}StrictOrderCheck && fieldIndex < previousFieldIndex {
			return totalN, fmt.Errorf("invalid order of fields (%d < %d): structure '%s' is out of order", fieldIndex, previousFieldIndex, structID)
		}
		missingFieldsByIndices[fieldIndex] = false

		var n int64
		switch structID {
 {{- range $index, $field := $struct.Fields }}
		case StructureID{{ $field.Struct.Name }}:
  {{- if $field.IsSlice }}
			var el {{ $field.Struct.TypeSpec.Name.String }}
			el.SetStructInfo(structInfo)
			n, err = el.ReadDataFrom(r)
			s.{{ $field.Name }} = append(s.{{ $field.Name }}, el)
  {{- else }}
			if fieldIndex == previousFieldIndex {
				return totalN, fmt.Errorf("field '{{ $field.Name }}' is not a slice, but multiple elements found")
			}
   {{- if $field.IsPointer }}
			s.{{ $field.Name }} = &{{ $field.Struct.TypeSpec.Name.String }}{}
   {{- end }}
			s.{{ $field.Name }}.SetStructInfo(structInfo)
			n, err = s.{{ $field.Name }}.ReadDataFrom(r) 
  {{- end }}
			if err != nil {
				return totalN, fmt.Errorf("unable to read field {{ $field.Name }} at %d: %w", totalN, err)
			}
 {{- end }}
		default:
			return totalN, fmt.Errorf("there is no field with structure ID '%s' in {{ $struct.Name }}", structInfo.ID)	
		}
    	{{- if $enableTracing }}
		fmt.Printf("%s parsed, TotalN is %d -> %d\n", structID, totalN, totalN + n){{- end}}
		totalN += n
		previousFieldIndex = fieldIndex
	}

	for fieldIndex, v := range missingFieldsByIndices {
		if v {
			return totalN, fmt.Errorf("field '%s' is missing", s.fieldNameByIndex(fieldIndex))
		}
	}

	return totalN, nil
}

{{- else }}

// ReadFrom reads the {{ $struct.Name }} from 'r' in format defined in the document #575623.
func (s *{{ $struct.Name }}) ReadFrom(r io.Reader) (int64, error) {
{{- if ne $struct.ElementStructID "" }}
	var totalN int64

	err := binary.Read(r, binary.LittleEndian, &s.StructInfo)
	if err != nil {
		return totalN, fmt.Errorf("unable to read structure info at %d: %w", totalN, err)
	}
	totalN += int64(binary.Size(s.StructInfo))

	n, err := s.ReadDataFrom(r)
	if err != nil {
		return totalN, fmt.Errorf("unable to read data: %w", err)
	}
	totalN += n

	return totalN, nil
}

// ReadDataFrom reads the {{ $struct.Name }} from 'r' excluding StructInfo,
// in format defined in the document #575623.
func (s *{{ $struct.Name }}) ReadDataFrom(r io.Reader) (int64, error) {
{{- end }}
	totalN := int64(0)
 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}

	// {{ $field.Name }} (ManifestFieldType: {{ $field.ManifestFieldType.String }})
    {{- if and ($enableTracing) (ne $fieldType "structInfo") }}
	fmt.Printf("{{ $struct.Name }}.{{ $field.Name }} (old TotalN is %d)\n", totalN){{- end}}
	{
  {{- if eq $fieldType "endValue" }}
		n, err := {{ $field.TypeStdSize}}, binary.Read(r, binary.LittleEndian, &s.{{ $field.Name }})
  {{- end }}
  {{- if eq $fieldType "structInfo" }}
		// ReadDataFrom does not read Struct, use ReadFrom for that.
  {{- end }}
  {{- if eq $fieldType "subStruct" }}
   {{- if $field.IsPointer }}
		s.{{ $field.Name }} = &{{ $field.ItemTypeName }}{}
   {{- end }}
		n, err := s.{{ $field.Name }}.ReadFrom(r)
  {{- end }}
  {{- if eq $fieldType "arrayStatic" }}
		n, err := {{ $field.TypeStdSize }}, binary.Read(r, binary.LittleEndian, s.{{ $field.Name }}[:])
  {{- end }}
  {{- if eq $fieldType "arrayDynamic" }}
   {{- if eq $field.CountValue "" }}
    	var size {{ $field.CountType }}
		err := binary.Read(r, binary.LittleEndian, &size)
		if err != nil {
			return totalN, fmt.Errorf("unable to the read size of field '{{ $field.Name }}': %w", err)	
		}
		totalN += int64(binary.Size(size))
   {{- else }}
		size := {{ $field.CountType }}(s.{{ $field.CountValue }})
   {{- end }}
		s.{{ $field.Name }} = make([]byte, size)
		n, err := len(s.{{ $field.Name }}), binary.Read(r, binary.LittleEndian, s.{{ $field.Name }})
  {{- end }}
  {{- if eq $fieldType "list" }}
    	var count {{ $field.CountType }}
		err := binary.Read(r, binary.LittleEndian, &count)
		if err != nil {
			return totalN, fmt.Errorf("unable to read the count for field '{{ $field.Name }}': %w", err)
		}
		totalN += int64(binary.Size(count))
		s.{{ $field.Name }} = make([]{{ $field.ItemTypeName }}, count)

		for idx := range s.{{ $field.Name }} {
			n, err := s.{{ $field.Name }}[idx].ReadFrom(r)
			if err != nil {
				return totalN, fmt.Errorf("unable to read field '{{ $field.Name }}[%d]': %w", idx, err)
			}
			totalN += int64(n)
		} 
  {{- end }}
  {{- if and (ne $fieldType "list") (ne $fieldType "structInfo") }}
		if err != nil {
			return totalN, fmt.Errorf("unable to read field '{{ $field.Name }}': %w", err)
		}
		totalN += int64(n)
  {{- end }}
	}
    {{- if and ($enableTracing) (ne $fieldType "structInfo") }}
	fmt.Printf("{{ $struct.Name }}.{{ $field.Name }} parsed, TotalN is %d)\n", totalN){{- end}}
 {{- end }}

	return totalN, nil
}

{{- end }}

// RehashRecursive calls Rehash (see below) recursively.
func (s *{{ $struct.Name }}) RehashRecursive() {
 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}
  {{- if or (eq $fieldType "subStruct") (eq $fieldType "structInfo") (eq $fieldType "element") }}
    {{- if $field.IsPointer }}                                                  
      	if s.{{ $field.Name }} != nil {                                         
      		s.{{ $field.Name }}.Rehash()                                    
        }                                                                       
    {{- else }}
	s.{{ $field.Name }}.Rehash()
	{{- end }}
  {{- end }}
 {{- end }}
	s.Rehash()
}

// Rehash sets values which are calculated automatically depending on the rest
// data. It is usually about the total size field of an element.
func (s *{{ $struct.Name }}) Rehash() {
 {{- if ne $struct.ElementStructInfoVar0 "" }}
	s.Variable0 = {{ $struct.ElementStructInfoVar0 }}
 {{- end }}
 {{- if ne $struct.ElementStructInfoVar1 "" }}
	s.ElementSize = {{ $struct.ElementStructInfoVar1 }}
 {{- end }}
 {{- range $index, $field := $struct.Fields }}
  {{- if ne $field.RehashValue "" }}
	s.{{ $field.Name }} = {{ $field.Type.Name }}(s.{{ $field.RehashValue }})
  {{- end }}
 {{- end }}
}

// WriteTo writes the {{ $struct.Name }} into 'w' in format defined in
// the document #575623.
func (s *{{ $struct.Name }}) WriteTo(w io.Writer) (int64, error) {
	totalN := int64(0)
	s.Rehash()

 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}

	// {{ $field.Name }} (ManifestFieldType: {{ $field.ManifestFieldType.String }})
	{{ if $field.IsPointer }}if s.{{ $field.Name }} != nil {{ end }}{
  {{- if eq $fieldType "endValue" }}
		n, err := {{ $field.TypeStdSize }}, binary.Write(w, binary.LittleEndian, &s.{{ $field.Name }})
  {{- end }}
  {{- if or (eq $fieldType "subStruct") (eq $fieldType "structInfo") (eq $fieldType "element") }}
		n, err := s.{{ $field.Name }}.WriteTo(w)
  {{- end }}
  {{- if eq $fieldType "arrayStatic" }}
		n, err := {{ $field.TypeStdSize }}, binary.Write(w, binary.LittleEndian, s.{{ $field.Name }}[:])
  {{- end }}
  {{- if eq $fieldType "arrayDynamic" }}
   {{- if eq $field.CountValue "" }}
		size := {{ $field.CountType }}(len(s.{{ $field.Name }}))
		err := binary.Write(w, binary.LittleEndian, size)
		if err != nil {
			return totalN, fmt.Errorf("unable to write the size of field '{{ $field.Name }}': %w", err)	
		}
		totalN += int64(binary.Size(size))
   {{- end }}
		n, err := len(s.{{ $field.Name }}), binary.Write(w, binary.LittleEndian, s.{{ $field.Name }})
  {{- end }}
  {{- if eq $fieldType "list" }}
    	count := {{ $field.CountType }}(len(s.{{ $field.Name }}))
		err := binary.Write(w, binary.LittleEndian, &count)
		if err != nil {
			return totalN, fmt.Errorf("unable to write the count for field '{{ $field.Name }}': %w", err)
		}
		totalN += int64(binary.Size(count))
  {{- end }}
  {{- if or (eq $fieldType "list") (eq $fieldType "elementList") }}
		for idx := range s.{{ $field.Name }} {
			n, err := s.{{ $field.Name }}[idx].WriteTo(w)
			if err != nil {
				return totalN, fmt.Errorf("unable to write field '{{ $field.Name }}[%d]': %w", idx, err)
			}
			totalN += int64(n)
		} 
  {{- end }}
  {{- if and (ne $fieldType "list") (ne $fieldType "elementList") }}
		if err != nil {
			return totalN, fmt.Errorf("unable to write field '{{ $field.Name }}': %w", err)
		}
		totalN += int64(n)
  {{- end }}
	}
 {{- end }}

	return totalN, nil
}

{{- range $index, $field := $struct.Fields }}
// {{ $field.Name }}Size returns the size in bytes of the value of field {{ $field.Name }} 
func (s *{{ $struct.Name }}) {{ $field.Name }}TotalSize() uint64 {
 {{- $fieldType := $field.ManifestFieldType.String }}
 {{- if or (eq $fieldType "endValue") (eq $fieldType "arrayStatic") }}
	return {{ $field.TypeStdSize }}
 {{- end }}
 {{- if or (eq $fieldType "subStruct") (eq $fieldType "structInfo") (eq $fieldType "element") }}
	return s.{{ $field.Name }}.TotalSize()
 {{- end }}
 {{- if or (eq $fieldType "list") (eq $fieldType "elementList") }}
	var size uint64
  {{- if and (eq $fieldType "list") (eq $field.CountValue "") }}
	size += uint64(binary.Size({{ $field.CountType }}(0)))
  {{- end }}
	for idx := range s.{{ $field.Name }} {
		size += s.{{ $field.Name }}[idx].TotalSize()
	}
	return size
 {{- end }}
 {{- if eq $fieldType "arrayDynamic" }}
  {{- if ne $field.CountValue "" }}
	return uint64(len(s.{{ $field.Name }}))
  {{- else }}
	size := uint64(binary.Size({{ $field.CountType }}(0)))
	size += uint64(len(s.{{ $field.Name }}))
	return size
  {{- end }}
 {{- end }}
}
{{- end }}

{{- range $index, $field := $struct.Fields }}
// {{ $field.Name }}Offset returns the offset in bytes of field {{ $field.Name }} 
func (s *{{ $struct.Name }}) {{ $field.Name }}Offset() uint64 {
 {{- if eq $index 0 }}
	return 0
 {{- else }}
  {{- $beforeField := index $struct.Fields (add $index -1) }}
	return s.{{ $beforeField.Name }}Offset() + s.{{ $beforeField.Name }}TotalSize() 
 {{- end }}
}
{{- end }}

// Size returns the total size of the {{ $struct.Name }}.  
func (s *{{ $struct.Name }}) TotalSize() uint64 {
	if s == nil {
		return 0
	}

	var size uint64
 {{- range $index, $field := $struct.Fields }}
	size += s.{{ $field.Name }}TotalSize()
 {{- end }}
	return size
}

// PrettyString returns the content of the structure in an easy-to-read format.
func (s *{{ $struct.Name }}) PrettyString(depth uint, withHeader bool) string {
	var lines []string
	if withHeader {
		lines = append(lines, pretty.Header(depth, {{ $struct.PrettyString | printf "%q" }}, s))
	}
	if s == nil {
		return strings.Join(lines, "\n")
	}
 {{- range $index, $field := $struct.Fields }}
  {{- $fieldType := $field.ManifestFieldType.String }}
	// ManifestFieldType is {{ $fieldType }}
  {{- if or (eq $fieldType "list") (eq $fieldType "elementList") }}
	lines = append(lines, pretty.Header(depth+1, fmt.Sprintf({{ printf "%s: Array of \"%s\" of length %%d" $field.Name $field.Struct.PrettyString | printf "%q"}}, len(s.{{ $field.Name }})), s.{{ $field.Name }}))
	for i := 0; i<len(s.{{ $field.Name }}); i++ {
		lines = append(lines, fmt.Sprintf("%sitem #%d: ", strings.Repeat("  ", int(depth+2)), i) + strings.TrimSpace(s.{{ $field.Name }}[i].PrettyString(depth+2, true)))
	}
	if depth < 1 {
		lines = append(lines, "")
	}
  {{- else }}
   {{- $fieldValue := ternary $field.IsPointer (printf "s.%s" $field.Name) (printf "&s.%s" $field.Name) }}
   {{- $prettyValue := ternary (ne $field.PrettyValue "") (printf "s.%s" $field.PrettyValue) $fieldValue }}
	lines = append(lines, pretty.SubValue(depth+1, {{ $field.PrettyString | printf "%q" }}, "", {{ $prettyValue }}))
  {{- end }} 
 {{- end }}
	if depth < 2 {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

{{- end }}

{{- range $index,$type := .BasicNamedTypes }}
// PrettyString returns the bits of the flags in an easy-to-read format.
func (flags {{ $type.Name }}) PrettyString(depth uint, withHeader bool) string {
 {{- if not (isNil ($type.MethodByName "String")) }}
	return flags.String()
 {{- else }}
	var lines []string
	if withHeader {
		lines = append(lines, pretty.Header(depth, {{ $type.PrettyString | printf "%q" }}, flags))
	}
  {{- range $index, $method := $type.Methods }}
   {{- if $method.ReturnsFlagValue }}
    {{- if eq $method.ReturnsTypeName "bool" }}
	if flags.{{ $method.Name }}() {
		lines = append(lines, pretty.SubValue(depth+1, "{{ $method.Name.Name | camelcaseToSentence }}", {{ $method.PrettyStringForResult true | printf "%q" }}, true))
	} else {
		lines = append(lines, pretty.SubValue(depth+1, "{{ $method.Name.Name | camelcaseToSentence }}", {{ $method.PrettyStringForResult false | printf "%q" }}, false))
	}
    {{- else }}
	lines = append(lines, pretty.SubValue(depth+1, "{{ $method.Name.Name | camelcaseToSentence }}", "", flags.{{ $method.Name }}()))
    {{- end }}
   {{- end }}
  {{- end }}
	return strings.Join(lines, "\n")
 {{- end }}
}
{{- end }}
`

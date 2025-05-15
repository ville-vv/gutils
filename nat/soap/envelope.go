package soap

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"reflect"
	"strings"
)

// ErrFault can be used as a target with errors.Is.
var ErrFault error = errors.New("xml fault")
var stringType = reflect.TypeOf("")
var _emptyStruct = &struct{}{}
var _ xml.Marshaler = &actionBuilder{}
var _ xml.Unmarshaler = &actionBuilder{}
var envOpen = []byte(xml.Header + `<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/"><s:Body>`)
var envClose = []byte(`</s:Body></s:Envelope>`)

type UPnPError struct {
	ErrorCode        string `xml:"errorCode"`
	ErrorDescription string `xml:"errorDescription"`
}

// FaultDetail carries XML-encoded application-specific Fault details.
type FaultDetail struct {
	//Raw       []byte    `xml:",innerxml"`
	UPnPError UPnPError `xml:"UPnPError"`
}

// Fault implements error, and contains SOAP fault information.
type Fault struct {
	Code   string      `xml:"faultcode"`
	String string      `xml:"faultstring"`
	Actor  string      `xml:"faultactor"`
	Detail FaultDetail `xml:"detail"`
}

func (fe *Fault) Error() string {
	if fe.Detail.UPnPError.ErrorCode != "" || fe.Detail.UPnPError.ErrorDescription != "" {
		return fmt.Sprintf("SOAP fault code=%s message=%s", fe.Detail.UPnPError.ErrorCode, fe.Detail.UPnPError.ErrorDescription)
	}
	return fmt.Sprintf("SOAP fault code=%s: message=%s", fe.Code, fe.String)
}

func (fe *Fault) Is(target error) bool {
	return target == ErrFault
}

type XMLName xml.Name

// actionBuilder wraps a SOAP action to be read or written as part of a SOAP envelope.
type actionBuilder struct {
	XMLName XMLName
	Space   string
	Args    interface{}
}

// newRecvAction creates a SOAP action for receiving arguments.
func newRecvAction(args interface{}) *actionBuilder {
	return &actionBuilder{Args: args}
}

// newSendAction creates a SOAP action for sending with the given namespace URL, action name, and arguments.
func newSendAction(serviceType string, args interface{}, actionName ...string) *actionBuilder {
	a := &actionBuilder{
		XMLName: XMLName{Space: serviceType},
		Space:   serviceType,
		Args:    args,
	}
	if len(actionName) > 0 {
		a.XMLName.Local = actionName[0]
	}
	return a
}

type ActionNameer interface {
	ActionName() string
}

func (a *actionBuilder) ActionName() string {
	if a.XMLName.Local != "" {
		return a.XMLName.Local
	}
	if an, ok := a.Args.(ActionNameer); ok {
		if name := an.ActionName(); name != "" {
			return name
		}
	}
	return a.getActionNameWithTag()
}

func (a *actionBuilder) getActionNameWithTag() string {
	if a.Args == nil || a.Args == _emptyStruct {
		return ""
	}
	tp := reflect.Indirect(reflect.ValueOf(a.Args)).Type()
	if tp.Kind() != reflect.Struct {
		return ""
	}
	for i := 0; i < tp.NumField(); i++ {
		field := tp.Field(i)
		// `soap:"action=xxx"`
		if tag := field.Tag.Get("action"); tag != "" {
			return tag
		}
	}
	return tp.Name()
}

func (a *actionBuilder) Namespace() string {
	if a.Space == "" {
		return a.XMLName.Space
	}
	return a.Space
}

// MarshalXML implements xml.Marshaller
// This is an implementation detail that allows packing elements inside the action element from the struct in `a.Args`.
func (a *actionBuilder) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if a.Args == nil {
		a.Args = _emptyStruct
	}
	vf := reflect.Indirect(reflect.ValueOf(a.Args))
	switch vf.Kind() {
	case reflect.Struct:
		return a.marshalStruct(e, vf)
	case reflect.Map:
		return a.marshalMap(e, vf)
	default:
		return fmt.Errorf("SOAP action does not support type as args: %s", vf.Type().String())
	}
}

func (a *actionBuilder) getStartElement() xml.StartElement {
	return xml.StartElement{
		Name: xml.Name{Local: "u:" + a.ActionName()},
		Attr: []xml.Attr{{
			Name:  xml.Name{Local: "xmlns:u"},
			Value: a.Namespace(),
		}},
	}
}

func (a *actionBuilder) marshalStruct(e *xml.Encoder, vl reflect.Value) error {
	return e.EncodeElement(a.Args, a.getStartElement())
}

func (a *actionBuilder) marshalMap(e *xml.Encoder, vl reflect.Value) error {
	tf := vl.Type()
	start := a.getStartElement()
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	kt := tf.Key()
	if kt.Kind() != reflect.String {
		return fmt.Errorf("SOAP action wants string as map key in args: %w", &xml.UnsupportedTypeError{Type: kt})
	}
	iter := vl.MapRange()
	for iter.Next() {
		k := iter.Key()
		ks := k.Convert(stringType).Interface().(string)
		v := iter.Value()
		ke := xml.StartElement{Name: xml.Name{Local: ks}}
		if err := e.EncodeElement(v.Interface(), ke); err != nil {
			return fmt.Errorf(
				"SOAP action error while encoding arg %q: %w", ks, err)
		}
	}
	return e.EncodeToken(xml.EndElement{Name: start.Name})
}

// UnmarshalXML implements xml.Unmarshaller
// This is an implementation detail that allows unpacking elements inside the
// action element into the struct in `a.Args`.
func (a *actionBuilder) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if a.Args == nil {
		a.Args = _emptyStruct
	}
	argsValue := reflect.Indirect(reflect.ValueOf(a.Args))
	argsType := argsValue.Type()
	switch argsType.Kind() {
	case reflect.Struct:
		return d.DecodeElement(a.Args, &start)
	case reflect.Map:
		keyType := argsType.Key()
		if keyType.Kind() != reflect.String {
			return fmt.Errorf("SOAP action wants string as map key in args: %w", &xml.UnsupportedTypeError{Type: keyType})
		}
		valueType := argsType.Elem()
		if valueType.Kind() == reflect.Interface {
			return fmt.Errorf("SOAP action wants a concrete type as map value in args: %w", &xml.UnsupportedTypeError{Type: valueType})
		}
		for {
			untypedToken, err := d.Token()
			if err != nil {
				return err
			}
			switch token := untypedToken.(type) {
			case xml.EndElement:
				return nil
			case xml.StartElement:
				if len(token.Attr) > 0 {
					return fmt.Errorf("SOAP action arg does not support attributes, got %v", token.Attr)
				}
				if token.Name.Space != "" {
					return fmt.Errorf("SOAP action arg does not support non-empty namespace, got %q", token.Name.Space)
				}
				key := reflect.ValueOf(token.Name.Local).Convert(keyType)
				value := reflect.New(valueType)
				if err := d.DecodeElement(value.Interface(), &token); err != nil {
					return fmt.Errorf("SOAP action arg %q errored while decoding: %w", key, err)
				}
				argsValue.SetMapIndex(key, reflect.Indirect(value))
			case xml.Comment:
			case xml.ProcInst:
				return fmt.Errorf("SOAP action args contained unexpected token %v", untypedToken)
			case xml.Directive:
				return fmt.Errorf("SOAP action args contained unexpected token %v", untypedToken)
			case xml.CharData:
				cd := string(token)
				if len(strings.TrimSpace(cd)) > 0 {
					return fmt.Errorf("SOAP action args contained stray text: %q", cd)
				}
			default:
				return fmt.Errorf("SOAP action found unknown XML token type: %T", untypedToken)
			}
		}
	default:
		return fmt.Errorf("SOAP action does not support type as args: %w", &xml.UnsupportedTypeError{Type: argsType})
	}
}

// encode marshals a SOAP envelope to the writer. Errors can be from the writer or XML encoding.
func (a *actionBuilder) encode(w io.Writer) error {
	enc := xml.NewEncoder(w)
	if err := enc.Encode(a); err != nil {
		return err
	}
	return enc.Flush()
}

type envelopeResp struct {
	XMLName       xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	EncodingStyle string   `xml:"http://schemas.xmlsoap.org/soap/envelope/ encodingStyle,attr"`
	Body          body     `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"`
}

type body struct {
	Fault  *Fault         `xml:"Fault"`
	Action *actionBuilder `xml:",any"`
}

func EncodeAction(w io.Writer, action *actionBuilder) error {
	_, err := w.Write(envOpen)
	if err != nil {
		return err
	}
	buf := bytes.NewBuffer(nil)
	if err = action.encode(buf); err != nil {
		return err
	}

	_, err = w.Write([]byte(html.UnescapeString(buf.String())))
	if err != nil {
		return err
	}
	_, err = w.Write(envClose)

	return err
}

// DecodeAction unmarshal a SOAP envelope from the reader. Errors can either be from the reader, XML decoding, or a *Fault.
func DecodeAction(r io.Reader, action *actionBuilder) error {
	env := envelopeResp{
		Body: body{
			Action: action,
		},
	}
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&env); err != nil {
		return err
	}
	if env.Body.Fault != nil {
		return env.Body.Fault
	}
	return nil
}

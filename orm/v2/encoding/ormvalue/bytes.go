package ormvalue

import (
	"bytes"
	"fmt"
	"io"

	"github.com/regen-network/regen-ledger/orm/v2/types/ormerrors"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type BytesCodec struct{}

func (b BytesCodec) FixedSize() int {
	panic("implement me")
}

func (b BytesCodec) Size(value protoreflect.Value) (int, error) {
	return bytesSize(value)
}

func bytesSize(value protoreflect.Value) (int, error) {
	bz := value.Bytes()
	n := len(bz)
	if n > 255 {
		return -1, ormerrors.BytesFieldTooLong
	}
	return n, nil
}

func (b BytesCodec) IsOrdered() bool {
	return false
}

func (b BytesCodec) Decode(r *bytes.Reader) (protoreflect.Value, error) {
	bz, err := io.ReadAll(r)
	return protoreflect.ValueOfBytes(bz), err
}

func (b BytesCodec) Encode(value protoreflect.Value, w io.Writer) error {
	_, err := w.Write(value.Bytes())
	return err
}

func (b BytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return bytes.Compare(v1.Bytes(), v2.Bytes())
}

func (b BytesCodec) IsEmpty(value protoreflect.Value) bool {
	return len(value.Bytes()) == 0
}

type NonTerminalBytesCodec struct{}

func (b NonTerminalBytesCodec) FixedSize() int {
	return -1
}

func (b NonTerminalBytesCodec) Size(value protoreflect.Value) (int, error) {
	n, err := bytesSize(value)
	return n + 1, err
}

func (b NonTerminalBytesCodec) IsOrdered() bool {
	return false
}

func (b NonTerminalBytesCodec) IsEmpty(value protoreflect.Value) bool {
	return len(value.Bytes()) == 0
}

func (b NonTerminalBytesCodec) Compare(v1, v2 protoreflect.Value) int {
	return bytes.Compare(v1.Bytes(), v2.Bytes())
}

func (b NonTerminalBytesCodec) Decode(r *bytes.Reader) (protoreflect.Value, error) {
	n, err := r.ReadByte()
	if err != nil {
		return protoreflect.Value{}, err
	}

	if n == 0 {
		return protoreflect.ValueOfBytes([]byte{}), nil
	}

	bz := make([]byte, n)
	_, err = r.Read(bz)
	return protoreflect.ValueOfBytes(bz), err
}

func (b NonTerminalBytesCodec) Encode(value protoreflect.Value, w io.Writer) error {
	bz := value.Bytes()
	n := len(bz)
	if n > 255 {
		return fmt.Errorf("can't encode a byte array longer than 255 bytes as an index part")
	}
	_, err := w.Write([]byte{byte(n)})
	if err != nil {
		return err
	}
	_, err = w.Write(bz)
	return err
}
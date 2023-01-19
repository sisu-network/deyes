package types

import (
	reflect "reflect"
	"testing"
	"unicode"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/require"
)

func TestAsset_Serialize(t *testing.T) {
	moduleId := uint32(70)
	assetId := uint32(80)
	fee := uint64(100)
	nonce := uint64(90)

	txMsg := TransactionMessage{
		ModuleID:        &moduleId,
		AssetID:         &assetId,
		Fee:             &fee,
		Asset:           []byte{19},
		Nonce:           &nonce,
		SenderPublicKey: []byte{100}, // TODO: check if this is correct
	}

	bz, err := proto.Marshal(&txMsg)
	require.Nil(t, err)

	tx := &TransactionMessage{}
	err = proto.Unmarshal(bz, tx)
	require.Nil(t, err)

	// Use reflection to compare all public fields
	types := reflect.TypeOf(*tx)
	values1 := reflect.ValueOf(*tx)
	values2 := reflect.ValueOf(txMsg)
	for i := 0; i < values1.NumField(); i++ {
		fieldName := types.Field(i).Name
		if !unicode.IsUpper(rune(fieldName[0])) {
			continue
		}

		f1 := values1.Field(i)
		if f1.Kind() == reflect.Pointer {
			require.Equal(t, f1.Elem().Interface(), values2.Field(i).Elem().Interface())
		} else {
			require.Equal(t, f1.Interface(), values2.Field(i).Interface())
		}
	}
}

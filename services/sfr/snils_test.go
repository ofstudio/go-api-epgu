package sfr

import (
	"errors"
	"fmt"
	"testing"
)

func ExampleSNILS_String() {
	snils := MustParseSNILS("715 398 174 20")
	fmt.Println(snils.String())
	// Output: 715-398-174 20
}

func ExampleSNILS_Number() {
	snils := MustParseSNILS("715 398 174 20")
	fmt.Println(snils.Number())
	// Output: 71539817420
}

func TestParseSNILS(t *testing.T) {
	tests := []struct {
		name    string
		number  string
		want    string
		wantErr error
	}{
		{name: "успешно пробелы", number: "276 488 905 42", want: "27648890542"},
		{name: "успешно дефисы", number: "002-064-585-96", want: "00206458596"},
		{name: "успешно дефисы и пробел", number: "774-112-048 81", want: "77411204881"},
		{name: "успешно слитно", number: "77241387515", want: "77241387515"},
		{name: "ошибка формат", number: "77 241 387 515", wantErr: ErrSNILSFormat},
		{name: "ошибка длина", number: "908 035 685", wantErr: ErrSNILSFormat},
		{name: "ошибка символы", number: "589*088*281*65", wantErr: ErrSNILSFormat},
		{name: "ошибка контрольная сумма", number: "200 746 095 00", wantErr: ErrSNILSCheck},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snils, err := ParseSNILS(tt.number)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseSNILS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := snils.Number()
			if got != tt.want {
				t.Errorf("ParseSNILS() got = %v, want %v", got, tt.want)
			}
		})
	}
}

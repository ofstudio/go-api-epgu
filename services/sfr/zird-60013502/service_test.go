package zird

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	apipgu "github.com/ofstudio/go-api-epgu"
	"github.com/ofstudio/go-api-epgu/services/sfr"
)

func exampleData() ZIRD {
	address := Address{ZipCode: "108809", Region: &AddressPart{Name: "Москва"}, City: &AddressPart{Name: "Москва"}, Settlement: &AddressPart{Name: "Толстопальцево"}, Street: &AddressPart{Name: "Ленина"}, House: "15", Flat: "106"}
	data := ZIRD{TOSFR: TerritorialOffice{Code: "087000", Name: "Отделение СФР по г. Москве и Московской области"}, Applicant: Applicant{FIO: sfr.FIO{LastName: "Борисов", FirstName: "Александр", PatronymicName: "Алексеевич"}, SNILS: sfr.MustParseSNILS("771-660-087 13"), BirthDate: sfr.NewDate(1972, 9, 10), Addresses: []AddressRus{{Type: AddressRegistration, Address: address}, {Type: AddressActual, Address: address}}, Phone: "89171034847", IdentityDoc: IdentityDoc{Code: "27", Name: "Паспорт гражданина РФ", Series: "51 24", Number: "110252", IssuedAt: sfr.NewDate(2020, 10, 10), IssuedBy: "ДЗЕРЖИНСКИМ РОВД Г. ЯРОСЛАВЛЯ", IssuerCode: "762-003"}}, Delivery: Delivery{Pension: PensionDelivery{Bank: Bank{Name: "Публичное акционерное общество Банк «ВТБ»", BIC: "044525187", AccountNumber: "40817810101006031825"}}}, FillingDate: sfr.NewDate(2026, 11, 10)}
	return data
}
func TestService(t *testing.T) {
	oldNow, oldGUID := nowFunc, guidFunc
	t.Cleanup(func() { nowFunc, guidFunc = oldNow, oldGUID })
	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	nowFunc = func() time.Time { return now }
	guidFunc = func() (string, error) { return "12345678-1234-1234-1234-123456789abc", nil }
	data := exampleData()
	data.FillingDate = sfr.Date{}
	s, err := NewService("45000000000", data)
	if err != nil {
		t.Fatal(err)
	}
	if s.ZIRD.FillingDate.Time != now || s.ExternalRegistrationDate.Time != now {
		t.Fatal("default dates")
	}
	meta := s.Meta()
	if meta.Region != "45000000000" || meta.ServiceCode != "60013502" || meta.TargetCode != "-60013502" {
		t.Fatal(meta)
	}
	dated, err := NewService("45", exampleData())
	if err != nil {
		t.Fatal(err)
	}
	if !dated.ZIRD.FillingDate.Equal(exampleData().FillingDate.Time) {
		t.Fatal("explicit date lost")
	}
	var logs bytes.Buffer
	s.WithDebug(logWriter{&logs})
	a, err := s.Archive(123456789)
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "123456789-archive" {
		t.Fatal(a.Name)
	}
	r, err := zip.NewReader(bytes.NewReader(a.Data), int64(len(a.Data)))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.File) != 2 {
		t.Fatal("archive contents")
	}
	for _, f := range r.File {
		rd, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(rd)
		rd.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(f.Name, "12345678-1234-1234-1234-123456789abc") {
			t.Fatal(f.Name)
		}
		schema := "ЗИРД_2025-05-01.xsd"
		if strings.HasPrefix(f.Name, "trans_") {
			schema = "schemas.xsd"
			for _, want := range []string{"<SNILS>77166008713</SNILS>", "<ExternalRegistrationNumber>123456789</ExternalRegistrationNumber>", "<ApplicationFileName>req_12345678-1234-1234-1234-123456789abc.xml</ApplicationFileName>"} {
				if !bytes.Contains(body, []byte(want)) {
					t.Fatal(want)
				}
			}
		} else {
			for _, want := range []string{"<Получатель>1</Получатель>", "<ns3:УчестьПрежнююДоставку>0</ns3:УчестьПрежнююДоставку>"} {
				if !bytes.Contains(body, []byte(want)) {
					t.Fatal(want)
				}
			}
			for _, bad := range []string{"МСЗ", "Почта", "Представитель", "СлужебнаяИнформация", "RegNum"} {
				if bytes.Contains(body, []byte(bad)) {
					t.Fatal(bad)
				}
			}
		}
		validateXML(t, schema, body)
	}
	if logs.Len() == 0 {
		t.Fatal("missing debug log")
	}
	s.WithDebug(nil)
	guidFunc = func() (string, error) { return "", errors.New("guid failure") }
	_, err = NewService("45", data)
	if !errors.Is(err, apipgu.ErrService) || !errors.Is(err, apipgu.ErrGUID) {
		t.Fatal(err)
	}
}

type logWriter struct{ *bytes.Buffer }

func (l logWriter) Print(v ...interface{}) { l.WriteString("logged") }
func validateXML(t *testing.T, schema string, body []byte) {
	t.Helper()
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal(body, &root); err != nil {
		t.Fatal(err)
	}
	binary, err := exec.LookPath("xmllint")
	if err != nil {
		t.Skip("xmllint unavailable")
	}
	p := filepath.Join(t.TempDir(), "document.xml")
	if err := os.WriteFile(p, body, 0600); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(binary, "--noout", "--schema", filepath.Join("xsd", schema), p).CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
}
func TestExamples(t *testing.T) {
	for name, schema := range map[string]string{"req-example.xml": "ЗИРД_2025-05-01.xsd", "trans-example.xml": "schemas.xsd"} {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("doc", name))
			if err != nil {
				t.Fatal(err)
			}
			validateXML(t, schema, b)
		})
	}
}

func TestFullAddressAndPaymentFile(t *testing.T) {
	data := exampleData()
	address := data.Applicant.Addresses[0].Address
	address.GUID = "12345678-1234-1234-1234-123456789abc"
	address.Area = &AddressPart{Name: "Округ", ShortName: "окр."}
	address.District = &AddressPart{Name: "Район"}
	address.CityDistrict = &AddressPart{Name: "Внутригородской район"}
	address.Building = "1"
	address.Estate = "2"
	address.Housing = "3"
	data.Applicant.Addresses = []AddressRus{{Type: AddressRegistration, Address: address}, {Type: AddressResidence, Address: address}, {Type: AddressActual, Address: address}}
	data.Applicant.DeliveryAddress = &address
	data.Delivery.Pension.RequestPaymentFile = true
	s, err := NewService("45", data)
	if err != nil {
		t.Fatal(err)
	}
	body, err := xml.Marshal(s.EDSFR)
	if err != nil {
		t.Fatal(err)
	}
	validateXML(t, "ЗИРД_2025-05-01.xsd", body)
	if !bytes.Contains(body, []byte("<ЗапросВыплатногоДела>true</ЗапросВыплатногоДела>")) {
		t.Fatal("missing payment file request")
	}
}

func TestDeliveryFlags(t *testing.T) {
	for _, keepPrevious := range []int{0, 1} {
		for _, requestPaymentFile := range []bool{false, true} {
			t.Run(fmt.Sprintf("keep_previous_%d_request_payment_file_%t", keepPrevious, requestPaymentFile), func(t *testing.T) {
				data := exampleData()
				data.Delivery.Pension.Bank.KeepPrevious = keepPrevious
				data.Delivery.Pension.RequestPaymentFile = requestPaymentFile
				srv, err := NewService("45000000000", data)
				if err != nil {
					t.Fatal(err)
				}
				check := func(body []byte) {
					t.Helper()
					for _, want := range []string{
						fmt.Sprintf("<ns3:УчестьПрежнююДоставку>%d</ns3:УчестьПрежнююДоставку>", keepPrevious),
						fmt.Sprintf("<ЗапросВыплатногоДела>%t</ЗапросВыплатногоДела>", requestPaymentFile),
					} {
						if !bytes.Contains(body, []byte(want)) {
							t.Fatalf("missing delivery flag: %s", want)
						}
					}
					validateXML(t, "ЗИРД_2025-05-01.xsd", body)
				}
				body, err := xml.Marshal(srv.EDSFR)
				if err != nil {
					t.Fatal(err)
				}
				check(body)

				archive, err := srv.Archive(123456789)
				if err != nil {
					t.Fatal(err)
				}
				reader, err := zip.NewReader(bytes.NewReader(archive.Data), int64(len(archive.Data)))
				if err != nil {
					t.Fatal(err)
				}
				for _, file := range reader.File {
					if file.Name != srv.Request.ApplicationFileName {
						continue
					}
					r, err := file.Open()
					if err != nil {
						t.Fatal(err)
					}
					body, err := io.ReadAll(r)
					r.Close()
					if err != nil {
						t.Fatal(err)
					}
					check(body)
					return
				}
				t.Fatal("application XML missing from archive")
			})
		}
	}
}

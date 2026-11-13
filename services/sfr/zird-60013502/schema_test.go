package zird_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ofstudio/go-api-epgu/services/sfr"
	zird "github.com/ofstudio/go-api-epgu/services/sfr/zird-60013502"
)

const (
	applicationSchema    = "ЗИРД_2025-05-01.xsd"
	transportSchema      = "zird-1.0.2.xsd"
	applicationNamespace = "http://пф.рф/ЗИРД/2025-05-01"
	typesNamespace       = "http://пф.рф/УТ/2025-01-01"
	bankNamespace        = "http://пф.рф/ВЗЛ/типы/2025-01-01"
)

// Все документы проверяются настоящим XSD-валидатором. Ошибки загрузки схем
// отделены от ошибок XML, чтобы не принять неисправный тест за дефект услуги.
type schemaValidator struct {
	binary  string
	schemas map[string]string
}

func newSchemaValidator(t *testing.T) *schemaValidator {
	t.Helper()
	binary, err := exec.LookPath("xmllint")
	if err != nil {
		t.Fatalf("для XSD-тестов требуется xmllint (libxml2): %v", err)
	}
	entries, err := os.ReadDir("xsd")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	schemas := make(map[string]string)
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".xsd" {
			schemas[entry.Name()] = filepath.Join(dir, fmt.Sprintf("schema-%02d.xsd", len(schemas)))
		}
	}
	// libxml2 не разрешает кириллические пути импортов в исходном комплекте.
	// Временные копии отличаются только именами файлов и schemaLocation.
	// Пространства имен, определения типов и ограничения не преобразуются.
	location := regexp.MustCompile(`\bschemaLocation\s*=\s*(?:"[^"]*"|'[^']*')`)
	for name, path := range schemas {
		data := readSchemaFile(t, filepath.Join("xsd", name))
		data = location.ReplaceAllFunc(data, func(attribute []byte) []byte {
			value := strings.TrimSpace(strings.SplitN(string(attribute), "=", 2)[1])
			imported := value[1 : len(value)-1]
			dest, ok := schemas[imported]
			if !ok {
				t.Fatalf("%s: импорт %q отсутствует в локальном комплекте XSD", name, imported)
			}
			return []byte(`schemaLocation="` + filepath.Base(dest) + `"`)
		})
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{applicationSchema, transportSchema} {
		if _, ok := schemas[name]; !ok {
			t.Fatalf("нет корневой схемы %s", name)
		}
	}
	return &schemaValidator{binary: binary, schemas: schemas}
}

func readSchemaFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func (v *schemaValidator) validate(t *testing.T, schema string, document []byte) (bool, string) {
	t.Helper()
	path, ok := v.schemas[schema]
	if !ok {
		t.Fatalf("неизвестная корневая схема %s", schema)
	}
	// Отрицательные фикстуры тоже должны быть синтаксически корректным XML.
	var root struct{ XMLName xml.Name }
	if err := xml.Unmarshal(document, &root); err != nil {
		t.Fatalf("ошибка XML в тесте: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, v.binary, "--nonet", "--noout", "--schema", path, "-")
	cmd.Stdin = bytes.NewReader(document)
	output, err := cmd.CombinedOutput()
	diagnostics := string(output)
	// Возвращаем в диагностике исходные названия XSD; номера строк сохранены.
	for name, staged := range v.schemas {
		diagnostics = strings.ReplaceAll(diagnostics, staged, filepath.Join("xsd", name))
	}
	if ctx.Err() != nil {
		t.Fatalf("тайм-аут XSD-валидатора: %v\n%s", ctx.Err(), diagnostics)
	}
	if err == nil {
		return true, diagnostics
	}
	var exit *exec.ExitError
	if errors.As(err, &exit) && exit.ExitCode() == 3 {
		return false, diagnostics // Схема загружена, XML не прошел валидацию.
	}
	t.Fatalf("не удалось выполнить XSD-проверку %s (ошибка окружения или компиляции схемы): %v\n%s", schema, err, diagnostics)
	return false, diagnostics
}

func (v *schemaValidator) requireValid(t *testing.T, schema string, document []byte) {
	t.Helper()
	if valid, diagnostics := v.validate(t, schema, document); !valid {
		t.Fatalf("XML не соответствует %s:\n%s\nПроверенный документ:\n%s", schema, diagnostics, document)
	}
}

func TestSchemaExamples(t *testing.T) {
	validator := newSchemaValidator(t)
	for _, tc := range []struct{ name, schema string }{
		{"req-example.xml", applicationSchema}, {"trans-example.xml", transportSchema},
	} {
		t.Run(tc.name, func(t *testing.T) {
			validator.requireValid(t, tc.schema, readSchemaFile(t, filepath.Join("doc", tc.name)))
		})
	}
}

// Данные вымышленного заявителя. Значения соответствуют форматам XSD;
// проверка действительности банковских реквизитов и справочников не выполняется.
func schemaApplication() zird.ZIRD {
	return zird.ZIRD{
		TOSFR: zird.TerritorialOffice{Code: "087", Name: "Отделение СФР по г. Москве и Московской области"},
		Applicant: zird.Applicant{
			FIO:   sfr.FIO{LastName: "Борисов", FirstName: "Александр", PatronymicName: "Алексеевич"},
			SNILS: sfr.MustParseSNILS("771-660-087 13"), BirthDate: sfr.NewDate(1972, 9, 10),
			Addresses: []zird.AddressRus{{Type: zird.AddressRegistration, Address: zird.Address{
				ZipCode: "108809", Region: &zird.AddressPart{Name: "Москва"}, City: &zird.AddressPart{Name: "Москва"},
				Settlement: &zird.AddressPart{Name: "Толстопальцево"}, Street: &zird.AddressPart{Name: "Ленина"}, House: "15", Flat: "106",
			}}},
			IdentityDoc: zird.IdentityDoc{Code: "27", Name: "Паспорт гражданина РФ", Series: "51 24", Number: "110252", IssuedAt: sfr.NewDate(2020, 10, 10), IssuedBy: "ДЗЕРЖИНСКИМ РОВД Г. ЯРОСЛАВЛЯ", IssuerCode: "762-003"},
			Phone:       "89171034847",
		},
		Delivery:    zird.Delivery{Pension: zird.PensionDelivery{Bank: zird.Bank{Name: "Банк «ВТБ»", BIC: "044525187", AccountNumber: "40817810101006031825", KeepPrevious: 1}}},
		FillingDate: sfr.NewDate(2026, 11, 13),
	}
}

func TestArchiveSchemas(t *testing.T) {
	validator := newSchemaValidator(t)
	cases := []struct {
		name   string
		change func(*zird.ZIRD)
	}{
		{"minimal", func(data *zird.ZIRD) {
			data.Applicant.FIO.PatronymicName = ""
			data.Applicant.Phone = ""
			data.Applicant.IdentityDoc.Series = ""
			data.Applicant.IdentityDoc.IssuerCode = ""
			data.Applicant.Addresses[0].Address = zird.Address{Region: &zird.AddressPart{Name: "Москва"}}
		}},
		{"one_address", func(*zird.ZIRD) {}},
		{"two_addresses", func(data *zird.ZIRD) {
			data.Applicant.Addresses = append(data.Applicant.Addresses, zird.AddressRus{Type: zird.AddressActual, Address: data.Applicant.Addresses[0].Address})
		}},
		{"three_addresses", func(data *zird.ZIRD) {
			address := data.Applicant.Addresses[0].Address
			data.Applicant.Addresses = append(data.Applicant.Addresses, zird.AddressRus{Type: zird.AddressResidence, Address: address}, zird.AddressRus{Type: zird.AddressActual, Address: address})
		}},
		{"full_address_and_delivery_address", func(data *zird.ZIRD) {
			address := &data.Applicant.Addresses[0].Address
			address.GUID = "12345678-1234-1234-1234-123456789abc"
			address.Region = &zird.AddressPart{Name: "Москва", ShortName: "г."}
			address.Area = &zird.AddressPart{Name: "Округ", ShortName: "окр."}
			address.District = &zird.AddressPart{Name: "Район", ShortName: "р-н"}
			address.CityDistrict = &zird.AddressPart{Name: "Внутригородской район", ShortName: "р-н"}
			address.Building = "1"
			address.Estate = "2"
			address.Housing = "3"
			delivery := *address
			delivery.House = "20"
			data.Applicant.DeliveryAddress = &delivery
		}},
	}
	for _, keep := range []int{0, 1} {
		for _, request := range []bool{false, true} {
			keep, request := keep, request
			cases = append(cases, struct {
				name   string
				change func(*zird.ZIRD)
			}{
				fmt.Sprintf("keep_%d_request_%t", keep, request), func(data *zird.ZIRD) {
					data.Delivery.Pension.Bank.KeepPrevious = keep
					data.Delivery.Pension.RequestPaymentFile = request
				},
			})
		}
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := schemaApplication()
			tc.change(&data)
			service, err := zird.NewService("45000000000", data)
			if err != nil {
				t.Fatal(err)
			}
			const orderID = 123456789
			archive, err := service.Archive(orderID)
			if err != nil {
				t.Fatal(err)
			}
			reader, err := zip.NewReader(bytes.NewReader(archive.Data), int64(len(archive.Data)))
			if err != nil {
				t.Fatal(err)
			}
			if len(reader.File) != 2 {
				t.Fatalf("ожидались два XML-файла, получено %d", len(reader.File))
			}
			var reqName, transName string
			var req, trans []byte
			for _, file := range reader.File {
				stream, err := file.Open()
				if err != nil {
					t.Fatal(err)
				}
				body, err := io.ReadAll(stream)
				closeErr := stream.Close()
				if err != nil {
					t.Fatal(err)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
				switch {
				case strings.HasPrefix(file.Name, "req_") && strings.HasSuffix(file.Name, ".xml"):
					if req != nil {
						t.Fatal("повторный файл заявления")
					}
					reqName, req = file.Name, body
				case strings.HasPrefix(file.Name, "trans_") && strings.HasSuffix(file.Name, ".xml"):
					if trans != nil {
						t.Fatal("повторный транспортный файл")
					}
					transName, trans = file.Name, body
				default:
					t.Fatalf("неожиданный файл %s", file.Name)
				}
			}
			if req == nil || trans == nil {
				t.Fatal("в архиве отсутствует заявление или конверт")
			}
			if strings.TrimPrefix(reqName, "req_") != strings.TrimPrefix(transName, "trans_") {
				t.Fatal("GUID файлов не совпадают")
			}
			t.Run("application", func(t *testing.T) { validator.requireValid(t, applicationSchema, req) })
			t.Run("transport", func(t *testing.T) { validator.requireValid(t, transportSchema, trans) })
			t.Run("values", func(t *testing.T) { checkArchiveValues(t, data, orderID, reqName, req, trans) })
		})
	}
}

func checkArchiveValues(t *testing.T, data zird.ZIRD, orderID int, reqName string, req, trans []byte) {
	t.Helper()
	// После XSD-проверки читаем значения без привязки к выбранным XML-префиксам.
	var application struct {
		Addresses []struct {
			Type int `xml:"ТипАдреса"`
		} `xml:"ЗИРД>Заявитель>Адрес"`
		DeliveryAddress *struct {
			House string `xml:"Дом"`
		} `xml:"ЗИРД>Заявитель>АдресДоставки"`
		SNILS              string `xml:"ЗИРД>Заявитель>СНИЛС"`
		Recipient          int    `xml:"ЗИРД>Доставка>Получатель"`
		RequestPaymentFile bool   `xml:"ЗИРД>Доставка>ЗДП>ЗапросВыплатногоДела"`
		KeepPrevious       int    `xml:"ЗИРД>Доставка>ЗДП>Банк>УчестьПрежнююДоставку"`
	}
	if err := xml.Unmarshal(req, &application); err != nil {
		t.Fatal(err)
	}
	if len(application.Addresses) != len(data.Applicant.Addresses) {
		t.Fatalf("число адресов не сохранилось: получено %d, ожидалось %d", len(application.Addresses), len(data.Applicant.Addresses))
	}
	for i, address := range application.Addresses {
		if address.Type != int(data.Applicant.Addresses[i].Type) {
			t.Fatalf("не сохранился тип адреса %d: %d", i, address.Type)
		}
	}
	if (application.DeliveryAddress == nil) != (data.Applicant.DeliveryAddress == nil) {
		t.Fatal("не сохранилось наличие отдельного адреса доставки")
	}
	if application.DeliveryAddress != nil && application.DeliveryAddress.House != data.Applicant.DeliveryAddress.House {
		t.Fatal("отдельный адрес доставки не соответствует входным данным")
	}
	if application.Recipient != 1 || application.RequestPaymentFile != data.Delivery.Pension.RequestPaymentFile || application.KeepPrevious != data.Delivery.Pension.Bank.KeepPrevious {
		t.Fatalf("значения доставки не сохранились: %+v", application)
	}
	var transport struct {
		SNILS               string `xml:"urn://iis.ecp.ru/zird/types/1.0.2 SNILS"`
		OrderID             string `xml:"ExternalRegistrationData>ExternalRegistrationNumber"`
		ApplicationFileName string `xml:"urn://iis.ecp.ru/zird/types/1.0.2 ApplicationFileName"`
	}
	if err := xml.Unmarshal(trans, &transport); err != nil {
		t.Fatal(err)
	}
	if application.SNILS != data.Applicant.SNILS.String() || transport.SNILS != data.Applicant.SNILS.Number() {
		t.Fatalf("СНИЛС заявления и конверта не соответствуют входным данным: %+v, %+v", application, transport)
	}
	if transport.OrderID != strconv.Itoa(orderID) || transport.ApplicationFileName != reqName {
		t.Fatalf("неверная связь конверта с заявлением: %+v", transport)
	}
}

// Эти проверки защищают от ложного успеха при запуске без нужной корневой
// схемы или без импортируемых ограничений. Портим контрольный XML, а не код услуги.
func TestSchemaRejectsInvalidXML(t *testing.T) {
	validator := newSchemaValidator(t)
	req := readSchemaFile(t, filepath.Join("doc", "req-example.xml"))
	trans := readSchemaFile(t, filepath.Join("doc", "trans-example.xml"))
	validator.requireValid(t, applicationSchema, req)
	validator.requireValid(t, transportSchema, trans)
	for _, tc := range []struct {
		name, schema         string
		source               []byte
		from, to, diagnostic string
	}{
		{"application_namespace", applicationSchema, req, applicationNamespace, "urn:wrong:zird", "ЭДСФР"},
		{"types_namespace", applicationSchema, req, typesNamespace, "urn:wrong:types", "ФИО"},
		{"bank_namespace", applicationSchema, req, bankNamespace, "urn:wrong:bank", "НаименованиеОрганизации"},
		{"required_element", applicationSchema, req, "<Получатель>1</Получатель>", "", "Получатель"},
		{"BIC", applicationSchema, req, "044525187", "BAD-BIC", "БИК"},
		{"account", applicationSchema, req, "40817810101006031825", "123", "НомерСчета"},
		{"boolean", applicationSchema, req, "<ns3:УчестьПрежнююДоставку>1</ns3:УчестьПрежнююДоставку>", "<ns3:УчестьПрежнююДоставку>2</ns3:УчестьПрежнююДоставку>", "УчестьПрежнююДоставку"},
		{"payment_file_boolean", applicationSchema, req, "<ЗапросВыплатногоДела>false</ЗапросВыплатногоДела>", "<ЗапросВыплатногоДела>2</ЗапросВыплатногоДела>", "ЗапросВыплатногоДела"},
		{"transport_SNILS", transportSchema, trans, "<SNILS>77166008713</SNILS>", "<SNILS>123</SNILS>", "SNILS"},
		{"transport_required_element", transportSchema, trans, "<SNILS>77166008713</SNILS>", "", "SNILS"},
		{"transport_namespace", transportSchema, trans, "urn://iis.ecp.ru/zird/1.0.2", "urn:wrong:transport", "Request"},
		{"fixed_AF", transportSchema, trans, `AF="ЗИРД_2025-05-01"`, `AF="wrong"`, "AF"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if bytes.Count(tc.source, []byte(tc.from)) != 1 {
				t.Fatalf("нет однозначной точки изменения контрольного XML: %s", tc.from)
			}
			mutated := bytes.Replace(tc.source, []byte(tc.from), []byte(tc.to), 1)
			valid, diagnostics := validator.validate(t, tc.schema, mutated)
			if valid {
				t.Fatalf("валидатор принял намеренно поврежденный XML: %s", tc.name)
			}
			if !strings.Contains(diagnostics, tc.diagnostic) {
				t.Fatalf("ошибка не связана с измененным полем %s:\n%s", tc.diagnostic, diagnostics)
			}
		})
	}
	t.Run("transport_sequence", func(t *testing.T) {
		sequence := regexp.MustCompile(`(?s)(<SNILS>[^<]*</SNILS>)(\s*)(<ExternalRegistrationData>.*?</ExternalRegistrationData>)`)
		if len(sequence.FindAll(trans, -1)) != 1 {
			t.Fatal("нет однозначной точки перестановки элементов")
		}
		mutated := sequence.ReplaceAll(trans, []byte(`${3}${2}${1}`))
		if valid, diagnostics := validator.validate(t, transportSchema, mutated); valid {
			t.Fatal("валидатор принял нарушенный порядок элементов")
		} else if !strings.Contains(diagnostics, "SNILS") {
			t.Fatalf("неожиданная диагностика:\n%s", diagnostics)
		}
	})
}

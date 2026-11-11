package zird

import (
	"encoding/xml"
	"fmt"
	"time"

	apipgu "github.com/ofstudio/go-api-epgu"
	"github.com/ofstudio/go-api-epgu/services/sfr"
	"github.com/ofstudio/go-api-epgu/utils"
)

const (
	FRGUTargetId = "364359127" // Идентификатор цели оказания госуслуги по ФРГУ
	ServiceCode  = "60013502"  // Идентификатор формы заявления
	TargetCode   = "-60013502" // Идентификатор цели
	MFCCode      = "API ЕПГУ"
)

var (
	errService    = fmt.Errorf("%w: %s", apipgu.ErrService, "60013502-sfr-zird")
	errXMLMarshal = fmt.Errorf("%w: %w", errService, apipgu.ErrXMLMarshal)
	errGUID       = fmt.Errorf("%w: %w", errService, apipgu.ErrGUID)
)

// Service - Услуга "Прием заявлений о доставке пенсии и иных социальных выплат"
type Service struct {
	EDSFR
	Request
	guid   string
	okato  string
	debug  bool
	logger utils.Logger
}

// NewService - конструктор [Service].
// Принимает код ОКАТО заявителя, а также данные заявления.
// В качестве ОКАТО можно использовать коды региона заявителя. Напр.: "92000000000".
//
// В случае ошибки возвращает цепочку из apipgu.ErrService и apipgu.ErrGUID.
func NewService(okato string, zird ZIRD) (*Service, error) {
	now := nowFunc()
	guid, err := guidFunc()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGUID, err)
	}

	if zird.FillingDate.IsZero() {
		zird.FillingDate.Time = now
	}

	return &Service{
		guid:  guid,
		okato: okato,
		EDSFR: EDSFR{
			Namespaces: edsfrNamespaces,
			ZIRD:       zird,
		},
		Request: Request{
			Namespaces:               requestNamespaces,
			AF:                       "ЗИРД_2025-05-01",
			SNILS:                    zird.Applicant.SNILS.Number(),
			ExternalRegistrationDate: sfr.Date{Time: now},
			MFCCode:                  MFCCode,
			FRGUTargetId:             FRGUTargetId,
		},
	}, nil
}

// WithDebug - включает логирование создаваемых XML-файлов и метаданных услуги.
// Формат лога:
//
//	>>> 60013502-sfr-zird: {имя файла}
//	...
//	{содержимое файла}
//	...
func (s *Service) WithDebug(logger utils.Logger) *Service {
	s.logger = logger
	s.debug = logger != nil
	return s
}

// Meta - возвращает метаданные услуги.
func (s *Service) Meta() apipgu.OrderMeta {
	meta := apipgu.OrderMeta{
		Region:      s.okato,
		ServiceCode: ServiceCode,
		TargetCode:  TargetCode,
	}
	s.logData("meta", meta.JSON())
	return meta
}

// Archive - возвращает архив с файлом заявления и транспортным файлом.
// В случае ошибки возвращает цепочку из apipgu.ErrService и следующих возможных ошибок:
//   - apipgu.ErrXMLMarshal - ошибка создания XML
//   - apipgu.ErrZip - ошибка создания zip-архива
func (s *Service) Archive(orderId int) (*apipgu.Archive, error) {
	var (
		archiveFileName = fmt.Sprintf("%d-archive", orderId)
		reqFileName     = fmt.Sprintf("req_%s.xml", s.guid)
		transFileName   = fmt.Sprintf("trans_%s.xml", s.guid)
	)

	s.Request.ApplicationFileName = reqFileName
	s.Request.ExternalRegistrationNumber = fmt.Sprintf("%d", orderId)

	reqXML, err := xml.MarshalIndent(s.EDSFR, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errXMLMarshal, err)
	}
	s.logData(reqFileName, reqXML)

	transXML, err := xml.MarshalIndent(s.Request, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errXMLMarshal, err)
	}
	s.logData(transFileName, transXML)

	reqFile := apipgu.ArchiveFile{Filename: reqFileName, Data: append([]byte(xml.Header), reqXML...)}
	transFile := apipgu.ArchiveFile{Filename: transFileName, Data: append([]byte(xml.Header), transXML...)}

	archive, err := apipgu.NewArchive(archiveFileName, reqFile, transFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errService, err)
	}

	return archive, nil
}

func (s *Service) logData(name string, data []byte) {
	if s.debug {
		s.logger.Print(fmt.Sprintf(">>> 60013502-sfr-zird: %s\n%s\n", name, string(data)))
	}
}

var (
	guidFunc = utils.GUID
	nowFunc  = time.Now
)

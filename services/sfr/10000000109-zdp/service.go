package zdp

import (
	"encoding/xml"
	"fmt"
	"time"

	apipgu "github.com/ofstudio/go-api-epgu"
	"github.com/ofstudio/go-api-epgu/services/sfr"
	"github.com/ofstudio/go-api-epgu/utils"
)

const (
	FRGUTargetId = "10002953957"  // Идентификатор цели оказания госуслуги по ФРГУ
	ServiceCode  = "10000000109"  // Идентификатор формы заявления
	TargetCode   = "-10000000109" // Идентификатор цели
	MFCCode      = "API ЕПГУ"
)

var (
	errService    = fmt.Errorf("%w: %s", apipgu.ErrService, "10000000109-sfr-zdp")
	errXMLMarshal = fmt.Errorf("%w: %w", errService, apipgu.ErrXMLMarshal)
	errGUID       = fmt.Errorf("%w: %w", errService, apipgu.ErrGUID)
)

// Service - Услуга "Доставка пенсии и социальных выплат ПФР"
type Service struct {
	EDPFR
	Request
	debug  bool
	logger utils.Logger
}

// NewService - конструктор [Service].
// Принимает коды ОКАТО и ОКТМО заявителя, а также данные заявления.
// В качестве ОКАТО и ОКТМО можно использовать коды региона заявителя. Напр.: "92000000000".
//
// В случае ошибки возвращает цепочку из apipgu.ErrService и apipgu.ErrGUID.
func NewService(okato, oktmo string, zdp ZDP) (*Service, error) {
	now := nowFunc()
	guid, err := guidFunc()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errGUID, err)
	}

	if zdp.FillingDate.IsZero() {
		zdp.FillingDate.Time = now
	}

	if zdp.DeliveryInfo.Date.IsZero() {
		zdp.DeliveryInfo.Date.Time = now
	}

	return &Service{
		EDPFR: EDPFR{
			Namespaces: edpfrNamespaces,
			ZDP:        zdp,
			ServiceInfo: ServiceInfo{
				GUID:            guid,
				DateTime:        sfr.DateTime{Time: now},
				ApplicationDate: sfr.Date{Time: now},
			},
		},
		Request: Request{
			Namespaces:               requestNamespaces,
			SNILS:                    zdp.Applicant.SNILS.Number(),
			ExternalRegistrationDate: sfr.Date{Time: now},
			OKATO:                    okato,
			OKTMO:                    oktmo,
			MFCCode:                  MFCCode,
			FRGUTargetId:             FRGUTargetId,
		},
	}, nil
}

// WithDebug - включает логирование создаваемых XML-файлов и метаданных услуги.
// Формат лога:
//
//	>>> 10000000109-sfr-zdp: {имя файла}
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
		Region:      s.Request.OKATO,
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
		reqFileName     = fmt.Sprintf("req_%s.xml", s.EDPFR.ServiceInfo.GUID)
		transFileName   = fmt.Sprintf("trans_%s.xml", s.EDPFR.ServiceInfo.GUID)
	)

	s.Request.ApplicationFileName = reqFileName
	s.Request.ExternalRegistrationNumber = fmt.Sprintf("%d", orderId)
	s.EDPFR.ServiceInfo.ExternalRegistrationNumber = s.Request.ExternalRegistrationNumber

	reqXML, err := xml.MarshalIndent(s.EDPFR, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errXMLMarshal, err)
	}
	s.logData(reqFileName, reqXML)

	transXML, err := xml.MarshalIndent(s.Request, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errXMLMarshal, err)
	}
	s.logData(transFileName, transXML)

	reqFile := apipgu.File{Filename: reqFileName, Data: append([]byte(xml.Header), reqXML...)}
	transFile := apipgu.File{Filename: transFileName, Data: append([]byte(xml.Header), transXML...)}

	archive, err := apipgu.NewArchive(archiveFileName, reqFile, transFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errService, err)
	}

	return archive, nil
}

func (s *Service) logData(name string, data []byte) {
	if s.debug {
		s.logger.Print(fmt.Sprintf(">>> 10000000109-sfr-zdp: %s\n%s\n", name, string(data)))
	}
}

var (
	guidFunc = utils.GUID
	nowFunc  = time.Now
)

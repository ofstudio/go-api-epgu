package main

import (
	"github.com/ofstudio/go-api-epgu/services/sfr"
	zdp "github.com/ofstudio/go-api-epgu/services/sfr/10000000109-zdp"
)

var (
	okato = "92000000000"
	oktmo = "92000000000"

	zdpData = zdp.ZDP{
		TOSFR: "Клиентская служба (на правах отдела) в Ново-Савиновском районе г.Казани",
		Applicant: zdp.Applicant{
			FIO: sfr.FIO{
				LastName:       "ИВАНОВ",
				FirstName:      "ИВАН",
				PatronymicName: "ИВАНОВИЧ",
			},
			Sex:       "М",
			BirthDate: sfr.NewDate(1952, 10, 18),
			SNILS:     sfr.MustParseSNILS("787-900-175 50"),
			BirthPlace: sfr.BirthPlace{
				Type: sfr.BirthPlaceSpecial,
				City: "Г. ОРЕЛ",
			},
			Citizenship: sfr.Citizenship{Type: sfr.CitizenshipRF},
			AddressFact: addressData,
			Phone:       "89123456789",
			IdentityDoc: sfr.IdentityDoc{
				Type:     sfr.IdentityDocPassportRF,
				Series:   "1234",
				Number:   "567890",
				IssuedAt: sfr.NewDate(2000, 10, 20),
				IssuedBy: "ФМС России",
			},
		},
		DeliveryInfo: zdp.DeliveryInfo{
			Location:      zdp.DeliveryBankOrHome,
			Method:        zdp.DeliveryBank,
			Recipient:     zdp.DeliveryMyself,
			Organisation:  "Филиал Банка «Южный» в г. Казани",
			AccountNumber: "40817000000000000001",
			Address:       *addressData,
		},
		Confirmation: 1,
	}

	addressData = sfr.NewAddressRus().
			WithZipCode("421001").
			WithRegion("Респ. Татарстан").
			WithCity("г. Казань").
			WithStreet("ул. Адоратского").
			WithHouse("д. 2А").
			WithHousing("корп. 1").
			WithFlat("кв. 1")
)

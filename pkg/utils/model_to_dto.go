package utils

import "pethelp-backend/internal/core/domain"

// ToSpecialistProfileDTO is a helper function to map the domain model to a DTO.
func ToSpecialistProfileDTO(specialistModel domain.Specialist) domain.SpecialistProfDTO {
	dto := domain.SpecialistProfDTO{
		ID:         specialistModel.ID,
		Email:      specialistModel.Email,
		IsActive:   specialistModel.IsActive,
		IsVerified: specialistModel.IsVerified,
	}

	if specialistModel.Name.Valid {
		dto.Name = &specialistModel.Name.String
	}
	if specialistModel.FamilyName.Valid {
		dto.FamilyName = &specialistModel.FamilyName.String
	}
	if specialistModel.Phone.Valid {
		dto.Phone = &specialistModel.Phone.String
	}
	if specialistModel.Bio.Valid {
		dto.Bio = &specialistModel.Bio.String
	}
	if specialistModel.Avatar.Valid {
		dto.AvatarURL = &specialistModel.Avatar.String
	}
	if len(specialistModel.ImageID) > 0 {
		for _, imgId := range specialistModel.ImageID {
			if imgId.Valid {
				url := imgId.String
				dto.PortfolioURLs = append(dto.PortfolioURLs, &url)
			}
		}
	}
	if specialistModel.Experience.Valid {
		val := specialistModel.Experience.Int32
		dto.Experience = &val
	}
	if specialistModel.Position.Valid {
		dto.Position = &specialistModel.Position.String
	}
	if specialistModel.Description.Valid {
		dto.Description = &specialistModel.Description.String
	}

	return dto
}

func ToSpecialistsDetailsDTO(specialistDetails domain.SpecialistDetails) domain.SpecialistDetailsDTO {
    
	baseProfile := ToSpecialistProfileDTO(specialistDetails.Specialist)

	dto := domain.SpecialistDetailsDTO{
		SpecialistProfDTO: baseProfile,
		ServicePriceDTO: domain.ServicePriceDTO{
			Service: specialistDetails.ServicePrice.Service.String,
			PricePerHour: specialistDetails.ServicePrice.PricePerHour,
			PricePerDay:  specialistDetails.ServicePrice.PricePerDay,
		},
	}

	return dto
}
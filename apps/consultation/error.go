package consultation

import "errors"

var (
	ErrInvalidTimeFormat    = errors.New("format waktu tidak valid")
	ErrTimeConflict         = errors.New("konflik waktu dengan konsultasi lain")
	ErrDoctorNotAvailable   = errors.New("dokter tidak tersedia")
	ErrConsultationPastDate = errors.New("tidak bisa membuat konsultasi untuk tanggal lalu")
	ErrInvalidFileType      = errors.New("tipe file tidak valid")
	ErrFileSizeExceeded     = errors.New("ukuran file melebihi batas 2MB")
	ErrDoctorNotFound       = errors.New("dokter tidak ditemukan")
)

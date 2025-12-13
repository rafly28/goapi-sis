package models

import (
	"time"
)

const (
	// Gender Enum
	GenderMale   = "Laki-laki"
	GenderFemale = "Perempuan"

	// Religion Enum
	ReligionIslam    = "Islam"
	ReligionKristen  = "Kristen"
	ReligionKatolik  = "Katolik"
	ReligionHindu    = "Hindu"
	ReligionBuddha   = "Buddha"
	ReligionKonghucu = "Konghucu"

	// Marital Status Enum
	MaritalSingle       = "Belum menikah"
	MaritalMarried      = "Menikah"
	MaritalSingleParent = "Single Parent"

	// Family Status Enum (Murid)
	FamilyKandung = "Anak Kandung"
	FamilyTiri    = "Anak Tiri"
	FamilyAngkat  = "Anak Angkat"
	FamilyLainnya = "Lainnya"

	RelationAyahK = "Ayah Kandung"
	RelationAyahA = "Ayah Angkat"
	RelationIbuK  = "Ibu Kandung"
	RelationIbuA  = "Ibu Angkat"
	RelationWali  = "Wali"

	// Job Type Enum (Orang Tua/Wali)
	JobIRT          = "Ibu Rumah Tangga"
	JobPNS          = "PNS"
	JobTNI_Polri    = "TNI/Polri"
	JobBUMN         = "BUMN/BUMD"
	JobSwasta       = "Karyawan Swasta"
	JobPetani       = "Petani/Pekebun"
	JobNelayan      = "Nelayan"
	JobWiraswasta   = "Wiraswasta"
	JobTidakBekerja = "Tidak Bekerja"
	JobLainnya      = "Lainnya"

	// Teacher Enums
	StatusPNS      = "PNS"
	StatusPPPK     = "PPPK"
	StatusKontrak  = "Kontrak Yayasan"
	StatusGuruTamu = "Guru Tamu"
	StatusHonorer  = "Honorer Sekolah"
	StatusLainnya  = "Lainnya"

	EduSMA = "SMA/SMK/MA"
	EduD1  = "D1"
	EduD2  = "D2"
	EduD3  = "D3"
	EduS1  = "S1"
	EduS2  = "S2"

	EmploymentGuruKelas  = "Guru Kelas"
	EmploymentGuruMatPel = "Guru Mata Pelajaran"
	EmploymentKepsek     = "Kepala Sekolah"
	EmploymentLainnya    = "Lainnya"
)

type User struct {
	UID          string    `json:"uid"`
	Username     string    `json:"username"`
	Pass         string    `json:"-"`
	RoleID       int       `json:"role_id"`
	RoleName     string    `json:"role_name,omitempty"`
	RefreshToken string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	RoleID    int       `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

type UserResponse struct {
	UID       string    `json:"uid"`
	Username  string    `json:"username"`
	RoleName  string    `json:"role_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type LoginCredentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type PaginationMeta struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalItems  int `json:"total_items"`
	Limit       int `json:"limit"`
}

type PaginatedUserResponse struct {
	Meta PaginationMeta `json:"meta"`
	Data []UserResponse `json:"data"`
}

type Person struct {
	UID           string `json:"uid"`            // PK & FK dari login_users
	FullName      string `json:"full_name"`      // VARCHAR(255)
	BirthDate     string `json:"birth_date"`     // Format YYYY-MM-DD
	NIK           string `json:"nik"`            // CHAR(16), NOT NULL
	Gender        string `json:"gender"`         // ENUM
	Religion      string `json:"religion"`       // ENUM
	MaritalStatus string `json:"marital_status"` // ENUM
	Address       string `json:"address"`        // TEXT
	PhoneNumber   string `json:"phone_number,omitempty"`
	Email         string `json:"email,omitempty"`
}

type RegisterBaseRequest struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	RoleID        int    `json:"role_id"`
	FullName      string `json:"full_name"`
	BirthDate     string `json:"birth_date"` // YYYY-MM-DD
	NIK           string `json:"nik"`
	Gender        string `json:"gender"`
	Religion      string `json:"religion"`
	MaritalStatus string `json:"marital_status"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
}

type UserProfileResponse struct {
	UID        string `json:"uid"`
	Username   string `json:"username"`
	RoleName   string `json:"role_name"`
	PersonData Person `json:"person_data"`
}

type StudentDetails struct {
	UID             string `json:"uid"`
	NIS             string `json:"nis,omitempty"`
	NISN            string `json:"nisn"`
	FamilyStatus    string `json:"family_status"`
	ChildOrder      int    `json:"child_order"`
	OriginSchool    string `json:"origin_school"`
	ReceivedClass   string `json:"received_class"`
	ReceivedDate    string `json:"received_date"` // YYYY-MM-DD
	FatherName      string `json:"father_name"`
	MotherName      string `json:"mother_name"`
	ParentAddress   string `json:"parent_address"`
	FatherJob       string `json:"father_job"`
	MotherJob       string `json:"mother_job"`
	GuardianName    string `json:"guardian_name,omitempty"`
	GuardianAddress string `json:"guardian_address,omitempty"`
	GuardianPhone   string `json:"guardian_phone,omitempty"`
	GuardianJob     string `json:"guardian_job,omitempty"`
}

type TeacherDetails struct {
	UID                 string `json:"uid"`
	NIP                 string `json:"nip,omitempty"`
	NUPTK               string `json:"nuptk,omitempty"`
	NRG                 string `json:"nrg,omitempty"`
	FunctionalPosition  string `json:"functional_position"`
	EmploymentStatus    string `json:"employment_status"`
	RankClass           string `json:"rank_class,omitempty"`
	HireDate            string `json:"hire_date"`
	SKAppointmentNumber string `json:"sk_appointment_number,omitempty"`
	EducatorCertNumber  string `json:"educator_cert_number,omitempty"`
	LastEducation       string `json:"last_education"`
	University          string `json:"university"`
	Major               string `json:"major"`
	GraduationYear      string `json:"graduation_year"`
	DiplomaNumber       string `json:"diploma_number,omitempty"`
}

type RegisterStudentRequest struct {
	RegisterBaseRequest // Mewarisi field Username, Password, FullName, dll
	StudentDetails      // Mewarisi field NISN, FatherName, dll
}

type RegisterTeacherRequest struct {
	RegisterBaseRequest
	TeacherDetails
}

type StudentProfileResponse struct {
	UID           string `json:"uid"`
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	BirthDate     string `json:"birth_date"`
	NIK           string `json:"nik"`
	Gender        string `json:"gender"`
	Religion      string `json:"religion"`
	MaritalStatus string `json:"marital_status"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number"`
	Email         string `json:"email"`
	NISN          string `json:"nisn"`
	NIS           string `json:"nis"`
	ReceivedDate  string `json:"received_date"`
	EntryYear     int    `json:"entry_year"`
}

type TeacherProfileResponse struct {
	UID                 string `json:"uid"`
	Username            string `json:"username"`
	FullName            string `json:"full_name"`
	BirthDate           string `json:"birth_date"`
	NIK                 string `json:"nik"`
	Gender              string `json:"gender"`
	Religion            string `json:"religion"`
	MaritalStatus       string `json:"marital_status"`
	Address             string `json:"address"`
	PhoneNumber         string `json:"phone_number"`
	Email               string `json:"email"`
	NIP                 string `json:"nip"`
	NUPTK               string `json:"nuptk"`
	NRG                 string `json:"nrg"`
	FunctionalPosition  string `json:"functional_position"`
	EmploymentStatus    string `json:"employment_status"`
	RankClass           string `json:"rank_class"`
	HireDate            string `json:"hire_date"`
	SKAppointmentNumber string `json:"sk_appointment_number"`
	EducatorCertNumber  string `json:"educator_cert_number"`
	LastEducation       string `json:"last_education"`
	University          string `json:"university"`
	Major               string `json:"major"`
	GraduationYear      string `json:"graduation_year"`
	DiplomaNumber       string `json:"diploma_number"`
	YearsOfServiceY     int    `json:"years_of_service_y"`
	YearsOfServiceM     int    `json:"years_of_service_m"`
}

type EditStudentRequest struct {
	// Data Person (Wajib)
	FullName      string `json:"full_name"`
	BirthDate     string `json:"birth_date"`
	Religion      string `json:"religion"`
	MaritalStatus string `json:"marital_status"`
	Address       string `json:"address"`
	PhoneNumber   string `json:"phone_number,omitempty"`
	Email         string `json:"email,omitempty"`
	NISN          string `json:"nisn"`
	NIS           string `json:"nis"`
	ReceivedDate  string `json:"received_date"`        // YYYY-MM-DD
	ParentUID     string `json:"parent_uid,omitempty"` // Jika ingin mengubah Parent
}

type EditTeacherRequest struct {
	// Data Person
	FullName            string `json:"full_name"`
	Religion            string `json:"religion"`
	MaritalStatus       string `json:"marital_status"`
	Address             string `json:"address"`
	PhoneNumber         string `json:"phone_number,omitempty"`
	Email               string `json:"email,omitempty"`
	NIP                 string `json:"nip"`
	FunctionalPosition  string `json:"functional_position"`
	NUPTK               string `json:"nuptk,omitempty"`
	NRG                 string `json:"nrg,omitempty"`
	EmploymentStatus    string `json:"employment_status"`
	RankClass           string `json:"rank_class,omitempty"`
	HireDate            string `json:"hire_date,omitempty"`
	SKAppointmentNumber string `json:"sk_appointment_number,omitempty"`
	EducatorCertNumber  string `json:"educator_cert_number,omitempty"`
	LastEducation       string `json:"last_education,omitempty"`
	University          string `json:"university,omitempty"`
	Major               string `json:"major,omitempty"`
	GraduationYear      string `json:"graduation_year,omitempty"`
	DiplomaNumber       string `json:"diploma_number,omitempty"`
}

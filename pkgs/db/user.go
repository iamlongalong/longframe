package db

type User struct {
	Model
	Name     string `gorm:"type:varchar(200)" json:"name"`
	Email    string `gorm:"type:varchar(200);uniqueIndex;" json:"email"`
	Password string `gorm:"type:varchar(200)" json:"password"`
	Phone    string `gorm:"type:varchar(200);uniqueIndex;" json:"phone"`
	Gender   string `gorm:"type:varchar(20);default:male"`
}

type SessUser struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	ID    string `json:"ID"`
}

func GetUser(uid string) (*User, error) {
	u := &User{}
	u.ID = uid

	if result := db.First(u); result.Error != nil {
		return nil, result.Error
	}

	return u, nil
}

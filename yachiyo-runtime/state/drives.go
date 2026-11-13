package state

type NamedDrive struct {
	Name  string
	Drive *Drive
}

type Drives struct {
	SocialDesire Drive
	Interest     Drive
}

func NewDrives() Drives {
	return Drives{
		SocialDesire: UrgencyDrive(Medium),
		Interest:     UrgencyDrive(Medium),
	}
}

func (d *Drives) All() []NamedDrive {
	return []NamedDrive{
		{Name: "SocialDesire", Drive: &d.SocialDesire},
		{Name: "Interest", Drive: &d.Interest},
	}
}

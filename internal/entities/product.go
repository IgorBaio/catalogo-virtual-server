package entities

/*
id,
  name,
  description,
  price,
  whatsappMessage,
  image = "https://github.com/shadcn.png",
*/

type Product struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
	IsActive    bool    `json:"isActive"`
	OwnerID     string  `json:"ownerId"`
}

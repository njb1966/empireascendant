package game

const (
	DefaultTurns      = 15
	DefaultMoney      = 10000
	DefaultMoneyBank  = 0
	DefaultPopulation = 20000
	DefaultFood       = 10000
	DefaultEnergy     = 300
)

type Empire struct {
	ID            string
	Username      string
	PasswordHash  string
	WorldName     string
	WorldNameKey  string
	EmpireName    string
	EmpireNameKey string
	TurnsLeft     int
	TurnDay       int64
	Money         int
	MoneyBank     int
	Population    int
	Food          int
	FoodStorage   int
	Energy        int
	ResearchPts   int
	BuildingPts   int
}

func ApplyDefaults(e *Empire, day int64) {
	e.TurnsLeft = DefaultTurns
	e.TurnDay = day
	e.Money = DefaultMoney
	e.MoneyBank = DefaultMoneyBank
	e.Population = DefaultPopulation
	e.Food = DefaultFood
	e.FoodStorage = 0
	e.Energy = DefaultEnergy
	e.ResearchPts = 0
	e.BuildingPts = 0
}

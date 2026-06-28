package game

const TravelSnapshotVersion = 1

type TravelSnapshot struct {
	Version   int             `json:"version"`
	GlobalID  string          `json:"global_id"`
	HomeNode  string          `json:"home_node"`
	FromNode  string          `json:"from_node"`
	Empire    Empire          `json:"empire"`
	Regions   []Region        `json:"regions"`
	Buildings Buildings       `json:"buildings"`
	Mines     []Mine          `json:"mines"`
	Tech      map[string]bool `json:"tech"`
	Military  Military        `json:"military"`
}

func NewTravelSnapshot(state EconomyState, globalID, homeNode, fromNode string) TravelSnapshot {
	regions := make([]Region, 0, len(state.Regions))
	for _, region := range state.Regions {
		regions = append(regions, region)
	}
	mines := make([]Mine, 0, len(state.Mines))
	for _, mine := range state.Mines {
		mines = append(mines, mine)
	}
	tech := make(map[string]bool, len(state.Tech))
	for key, unlocked := range state.Tech {
		tech[key] = unlocked
	}
	return TravelSnapshot{
		Version:   TravelSnapshotVersion,
		GlobalID:  globalID,
		HomeNode:  homeNode,
		FromNode:  fromNode,
		Empire:    state.Empire,
		Regions:   regions,
		Buildings: state.Buildings,
		Mines:     mines,
		Tech:      tech,
		Military:  state.Military,
	}
}

func (s TravelSnapshot) EconomyState() EconomyState {
	regions := make(map[string]Region, len(s.Regions))
	for _, region := range s.Regions {
		regions[region.Type] = region
	}
	mines := make(map[string]Mine, len(s.Mines))
	for _, mine := range s.Mines {
		mines[mine.Type] = mine
	}
	tech := make(map[string]bool, len(s.Tech))
	for key, unlocked := range s.Tech {
		tech[key] = unlocked
	}
	return EconomyState{
		Empire:    s.Empire,
		Regions:   regions,
		Buildings: s.Buildings,
		Mines:     mines,
		Tech:      tech,
		Military:  s.Military,
	}
}

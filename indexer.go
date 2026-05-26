package indexer

import (
	"sync"
	"time"
)

// Menambahkan Mutex agar aman dari Race Condition saat API diakses bersamaan
type IndexerState struct {
	mu                sync.RWMutex
	LastIndexedHeight uint64                     `json:"last_indexed_height"`
	TPS               float64                    `json:"tps"`
	Blocks            []BlockRecord              `json:"blocks"`
	Transactions      []TxRecord                 `json:"transactions"`
	Addresses         map[string]AddressRecord   `json:"addresses"`
	Contracts         map[string]ContractRecord  `json:"contracts"`
	Tokens            map[string]TokenRecord     `json:"tokens"`
	Validators        map[string]ValidatorRecord `json:"validators"`
	DailyStats        map[string]DailyStat       `json:"daily_stats"` // UNTUK GRAFIK CHART
	UpdatedAt         int64                      `json:"updated_at"`
}

// 1. TAMBAHAN STRUCT BARU
type DailyStat struct {
	Date     string `json:"date"` // Format: YYYY-MM-DD
	TxCount  int    `json:"tx_count"`
	GasSpent uint64 `json:"gas_spent"`
}

type AddressRecord struct {
	Address       string            `json:"address"`
	Balance       string            `json:"balance"` // Saldo koin utama (ANR)
	Stake         string            `json:"stake"`
	TxCount       int               `json:"tx_count"`
	TokenBalances map[string]string `json:"token_balances"` // PORTOFOLIO TOKEN: TokenMint -> Saldo
	UpdatedAt     int64             `json:"updated_at"`
}

type ValidatorRecord struct {
	Address      string  `json:"address"`
	Power        string  `json:"power"`
	Active       bool    `json:"active"`
	UptimeBps    int     `json:"uptime_bps"`
	MissedBlocks int     `json:"missed_blocks"`
	Slashed      bool    `json:"slashed"`
	Commission   float64 `json:"commission"` // Fee Validator (misal 5.0%)
	UpdatedAt    int64   `json:"updated_at"`
}

// Struct lainnya tetap sama
type BlockRecord struct {
	Height    uint64 `json:"height"`
	Hash      string `json:"hash"`
	Time      int64  `json:"time"`
	Proposer  string `json:"proposer"`
	TxCount   int    `json:"tx_count"`
	GasUsed   uint64 `json:"gas_used"`
	StateRoot string `json:"state_root"`
}

type TxRecord struct {
	Hash   string `json:"hash"`
	Height uint64 `json:"height"`
	Type   string `json:"type"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
	Fee    string `json:"fee"`
	Status string `json:"status"`
	Time   int64  `json:"time"`
}

type ContractRecord struct {
	Address   string `json:"address"`
	Name      string `json:"name"`
	Standard  string `json:"standard"`
	Verified  bool   `json:"verified"`
	CodeHash  string `json:"code_hash"`
	UpdatedAt int64  `json:"updated_at"`
}

type TokenRecord struct {
	Mint          string `json:"mint"`
	Name          string `json:"name"`
	Symbol        string `json:"symbol"`
	Decimals      uint8  `json:"decimals"`
	LogoURI       string `json:"logo_uri"`
	HolderCount   int    `json:"holder_count"`
	TransferCount int    `json:"transfer_count"`
	UpdatedAt     int64  `json:"updated_at"`
}

func NewState() *IndexerState {
	now := time.Now().Unix()
	return &IndexerState{
		Blocks:            []BlockRecord{},
		Transactions:      []TxRecord{},
		Addresses:         make(map[string]AddressRecord),
		Contracts:         make(map[string]ContractRecord),
		Tokens:            make(map[string]TokenRecord),
		Validators:        make(map[string]ValidatorRecord),
		DailyStats:        make(map[string]DailyStat),
		UpdatedAt:         now,
		LastIndexedHeight: 0,
		TPS:               0.0,
	}
}

// Gunakan Mutex untuk thread-safety
func (s *IndexerState) AddBlock(b BlockRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.Blocks = append(s.Blocks, b)
	if b.Height > s.LastIndexedHeight {
		s.LastIndexedHeight = b.Height
	}
	
	// Update statistik harian untuk chart
	dateStr := time.Unix(b.Time, 0).Format("2006-01-02")
	stat := s.DailyStats[dateStr]
	stat.Date = dateStr
	stat.TxCount += b.TxCount
	stat.GasSpent += b.GasUsed
	s.DailyStats[dateStr] = stat

	s.UpdatedAt = time.Now().Unix()
}

// Contoh fungsi untuk mendapatkan estimasi gas dari blok terakhir
func (s *IndexerState) GetGasEstimation() map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Logika: ambil rata-rata gas dari 10 blok terakhir
	// Jika belum ada cukup blok, berikan nilai default
	var totalGas uint64
	count := 0
	for i := len(s.Blocks) - 1; i >= 0 && count < 10; i-- {
		totalGas += s.Blocks[i].GasUsed
		count++
	}

	avgGas := 1.2 // Default base Gwei
	if count > 0 {
		avgGas = float64(totalGas/uint64(count)) / 1000000 // Sesuaikan rumus dengan satuan Anda
	}

	return map[string]float64{
		"low":     avgGas * 0.8,
		"average": avgGas,
		"high":    avgGas * 1.5,
	}
}

// Contoh HTTP Handler di Go
func GasEstimateHandler(w http.ResponseWriter, r *http.Request) {
    data := state.GetGasEstimation() // Ambil dari IndexerState
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w.Encode(data))
}

// Tambahkan s.mu.Lock() dan defer s.mu.Unlock() pada metode AddTx, UpsertAddress dll seperti fungsi AddBlock di atas...
package block

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"jhblockchain/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
)

const MINING_DIFFICULT = 3
const MINING_ACCOUNT_ADDRESS = "JOHNHAI BLOCKCHAIN"
const MINING_REWARD = 5000
const MINING_TIMER_SEC = 10
const (
	//以下参数可以添加到启动参数
	BLOCKCHAIN_PORT_RANGE_START      = 5000
	BLOCKCHAIN_PORT_RANGE_END        = 5003
	NEIGHBOR_IP_RANGE_START          = 0
	NEIGHBOR_IP_RANGE_END            = 0
	BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC = 10
)

type Block struct {
	nonce        int
	previousHash [32]byte
	timestamp    int64
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, txs []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = txs
	return b
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Transactions() []*Transaction {
	return b.transactions
}

func (b *Block) Print() {
	log.Printf("%-15v:%30d\n", "timestamp", b.timestamp)
	//fmt.Printf("timestamp       %d\n", b.timestamp)
	log.Printf("%-15v:%30d\n", "nonce", b.nonce)
	log.Printf("%-15v:%30x\n", "previous_hash", b.previousHash)
	//log.Printf("%-15v:%30s\n", "transactions", b.transactions)
	for _, t := range b.transactions {
		t.Print()
	}
}

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	port              uint16
	mux               sync.Mutex

	neighbors    []string
	muxNeighbors sync.Mutex
}

// 新建一条链的第一个区块
// NewBlockchain(blockchainAddress string) *Blockchain
// 函数定义了一个创建区块链的方法，它接收一个字符串类型的参数 blockchainAddress，
// 它返回一个区块链类型的指针。在函数内部，它创建一个区块链对象并为其设置地址，
// 然后创建一个创世块并将其添加到区块链中，最后返回区块链对象。
func NewBlockchain(blockchainAddress string, port uint16) *Blockchain {
	bc := new(Blockchain)
	b := &Block{}
	bc.CreateBlock(0, b.Hash()) //创世纪块
	bc.blockchainAddress = blockchainAddress

	bc.port = port
	return bc
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) Run() {

	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
	bc.StartMining()
}

func (bc *Blockchain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(
		utils.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)

	color.Blue("邻居节点：%v", bc.neighbors)
}

func (bc *Blockchain) SyncNeighbors() {
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *Blockchain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	_ = time.AfterFunc(time.Second*BLOCKCHIN_NEIGHBOR_SYNC_TIME_SEC, bc.StartSyncNeighbors)
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) ClearTransactionPool() {
	bc.transactionPool = bc.transactionPool[:0]
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Blocks *[]*Block `json:"chain"`
	}{
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

// (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block
//  函数是在区块链上创建新的区块，它接收两个参数：一个int类型的nonce和一个字节数组类型的 previousHash，
//  返回一个区块类型的指针。在函数内部，它使用传入的参数来创建一个新的区块，
//  然后将该区块添加到区块链的链上，并清空交易池。

func (bc *Blockchain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}

	// 删除其他节点的交易
	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, _ := http.NewRequest("DELETE", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
	return b
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		color.Green("%s BLOCK %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	color.Yellow("%s\n\n\n", strings.Repeat("*", 50))
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

func (b *Block) MarshalJSON() ([]byte, error) {

	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int            `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	return nil
}

func (bc *Blockchain) LastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) AddTransaction(
	sender string,
	recipient string,
	value int64,
	senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	//如果是挖矿得到的奖励交易，不验证
	if sender == MINING_ACCOUNT_ADDRESS {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}

	// 判断有没有足够的余额
	// log.Printf("transaction.go sender:%s  account=%d", sender, bc.CalculateTotalAmount(sender))
	// if bc.CalculateTotalAmount(sender) <= uint64(value) {
	// 	log.Printf("ERROR: %s ，你的钱包里没有足够的钱", sender)
	// 	return false
	// }

	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {

		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: 验证交易")
	}
	return false

}

func (bc *Blockchain) CreateTransaction(sender string, recipient string, value uint64,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransacted := bc.AddTransaction(sender, recipient, int64(value), senderPublicKey, s)

	if isTransacted {
		for _, n := range bc.neighbors {
			publicKeyStr := fmt.Sprintf("%064x%064x", senderPublicKey.X.Bytes(),
				senderPublicKey.Y.Bytes())
			signatureStr := s.String()
			bt := &TransactionRequest{
				&sender, &recipient, &publicKeyStr, &value, &signatureStr}
			m, _ := json.Marshal(bt)
			buf := bytes.NewBuffer(m)
			endpoint := fmt.Sprintf("http://%s/transactions", n)
			client := &http.Client{}
			req, _ := http.NewRequest("PUT", endpoint, buf)
			resp, _ := client.Do(req)
			log.Printf("   **  **  **  CreateTransaction : %v", resp)
		}
	}

	return isTransacted
}

func (bc *Blockchain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderAddress,
				t.receiveAddress,
				t.value))
	}
	return transactions
}

func (bc *Blockchain) ValidProof(nonce int,
	previousHash [32]byte,
	transactions []*Transaction,
	difficulty int,
) bool {
	zeros := strings.Repeat("0", difficulty)
	//zeros := "1234"
	// tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions, timestamp: time.Now().UnixNano()}
	tmpBlock := Block{nonce: nonce, previousHash: previousHash, transactions: transactions}

	//log.Printf("tmpBlock%+v", tmpBlock)
	tmpHashStr := fmt.Sprintf("%x", tmpBlock.Hash())
	//log.Println("guessHashStr", tmpHashStr)
	return tmpHashStr[:difficulty] == zeros
}

func (bc *Blockchain) ProofOfWork() int {
	transactions := bc.CopyTransactionPool() //选择交易？控制交易数量？
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	begin := time.Now()
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULT) {
		nonce += 1
	}

	end := time.Now()
	//log.Printf("POW spend Time:%f", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%f Second", end.Sub(begin).Seconds())
	log.Printf("POW spend Time:%s", end.Sub(begin))

	return nonce
}

// 将交易池的交易打包
func (bc *Blockchain) Mining() bool {
	bc.mux.Lock()

	defer bc.mux.Unlock()

	// 此处判断交易池是否有交易，你可以不判断，打包无交易区块
	if len(bc.transactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINING_ACCOUNT_ADDRESS, bc.blockchainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}

	return true
}

func (bc *Blockchain) CalculateTotalAmount(accountAddress string) uint64 {
	var totalAmount uint64 = 0
	for _, _chain := range bc.chain {
		for _, _tx := range _chain.transactions {
			if accountAddress == _tx.receiveAddress {
				totalAmount = totalAmount + uint64(_tx.value)
			}
			if accountAddress == _tx.senderAddress {
				totalAmount = totalAmount - uint64(_tx.value)
			}
		}
	}
	return totalAmount
}

func (bc *Blockchain) StartMining() {
	bc.Mining()
	// 使用time.AfterFunc函数创建了一个定时器，它在指定的时间间隔后执行bc.StartMining函数（自己调用自己）。
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
	color.Yellow("minetime: %v\n", time.Now())
}

type AmountResponse struct {
	Amount uint64 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount uint64 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

type Transaction struct {
	senderAddress  string
	receiveAddress string
	value          int64
}

func NewTransaction(sender string, receive string, value int64) *Transaction {
	t := Transaction{sender, receive, value}
	return &t
}

func (bc *Blockchain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (t *Transaction) Print() {
	color.Red("%s\n", strings.Repeat("~", 30))
	color.Cyan("发送地址             %s\n", t.senderAddress)
	color.Cyan("接受地址             %s\n", t.receiveAddress)
	color.Cyan("金额                 %d\n", t.value)

}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string `json:"sender_blockchain_address"`
		Recipient string `json:"recipient_blockchain_address"`
		Value     int64  `json:"value"`
	}{
		Sender:    t.senderAddress,
		Recipient: t.receiveAddress,
		Value:     t.value,
	})
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	v := &struct {
		Sender    *string `json:"sender_blockchain_address"`
		Recipient *string `json:"recipient_blockchain_address"`
		Value     *int64  `json:"value"`
	}{
		Sender:    &t.senderAddress,
		Recipient: &t.receiveAddress,
		Value:     &t.value,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) ValidChain(chain []*Block) bool {
	preBlock := chain[0]
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		if b.previousHash != preBlock.Hash() {
			return false
		}

		if !bc.ValidProof(b.Nonce(), b.PreviousHash(), b.Transactions(), MINING_DIFFICULT) {
			return false
		}

		preBlock = b
		currentIndex += 1
	}
	return true
}

func (bc *Blockchain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLength := len(bc.chain)

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, err := http.Get(endpoint)
		if err != nil {
			color.Red("                 错误 ：ResolveConflicts GET请求")
			return false
		} else {
			color.Green("                正确 ：ResolveConflicts  GET请求")
		}
		if resp.StatusCode == 200 {
			var bcResp Blockchain
			decoder := json.NewDecoder(resp.Body)
			err1 := decoder.Decode(&bcResp)

			if err1 != nil {
				color.Red("                 错误 ：ResolveConflicts Decode")
				return false
			} else {
				color.Green("                正确 ：ResolveConflicts  Decode")
			}

			chain := bcResp.Chain()
			color.Cyan("   ResolveConflicts   chain len:%d ", len(chain))
			if len(chain) > maxLength && bc.ValidChain(chain) {
				maxLength = len(chain)
				longestChain = chain
			}
		}
	}

	color.Cyan("   ResolveConflicts   longestChain len:%d ", len(longestChain))

	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resovle confilicts replaced")
		return true
	}
	log.Printf("Resovle conflicts not replaced")
	return false
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	SenderPublicKey            *string `json:"sender_public_key"`
	Value                      *uint64 `json:"value"`
	Signature                  *string `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockchainAddress == nil ||
		tr.RecipientBlockchainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}

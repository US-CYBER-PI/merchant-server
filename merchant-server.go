package main

import (
	_interface "MerchantServer/interface"
	"MerchantServer/repositories"
	"MerchantServer/utils"
	"encoding/json"
	"fmt"
	qiwiSdk "github.com/US-CYBER-PI/qiwi-bill-paymentsgo-sdk/src"
	qiwiSdkModels "github.com/US-CYBER-PI/qiwi-bill-paymentsgo-sdk/src/Models"
	qiwiSdkSubModels "github.com/US-CYBER-PI/qiwi-bill-paymentsgo-sdk/src/Models/Ets"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	pgUser            = "secret"
	pgPassword        = "secret"
	pgHost            = "localhost"
	pgPort            = "5432"
	pgDB              = "jwt"
	pgPaymentTable    = "payments"
	pgUserTable       = "users"
	pgTokenTable      = "tokens"
	port              = 9096
	hmacSecret        = []byte("c3bd7d88edb4fa1817abb11702158924384f7933e5facfd707dc1d1429af9931")
	qiwiToken         = "8f69ff16-d505-1ed1-84e3-f677467a5c23"
	qiwiSiteId        = "sa4kjn-12"
	jwtManager        *utils.JwtManager
	qiwiClient        qiwiSdk.Client
	paymentRepository _interface.PaymentRepository
	userRepository    _interface.UserRepository
)

func init() {

	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file")
	}

	if os.Getenv("PG_USER") != "" {
		pgUser = os.Getenv("PG_USER")
	}

	if os.Getenv("PG_PASSWORD") != "" {
		pgPassword = os.Getenv("PG_PASSWORD")
	}

	if os.Getenv("PG_HOST") != "" {
		pgHost = os.Getenv("PG_HOST")
	}

	if os.Getenv("PG_PORT") != "" {
		pgPort = os.Getenv("PG_PORT")
	}

	if os.Getenv("PG_DB") != "" {
		pgDB = os.Getenv("PG_DB")
	}

	if os.Getenv("PG_PAYMENT_TABLE") != "" {
		pgPaymentTable = os.Getenv("PG_PAYMENT_TABLE")
	}

	if os.Getenv("PORT") != "" {
		port, _ = strconv.Atoi(os.Getenv("PORT"))
	}

	if os.Getenv("HMAC_SECRET") != "" {
		hmacSecret = []byte(os.Getenv("HMAC_SECRET"))
	}

	if os.Getenv("QIWI_TOKEN") != "" {
		qiwiToken = os.Getenv("QIWI_TOKEN")
	}

	if os.Getenv("QIWI_SITE_ID") != "" {
		qiwiSiteId = os.Getenv("QIWI_SITE_ID")
	}
}

func main() {

	var err error
	jwtManager = utils.NewJwtManager(hmacSecret)

	qiwiClient = *qiwiSdk.NewClient(qiwiToken, "https://api.qiwi.com/partner", qiwiSiteId)

	paymentRepository, err = repositories.NewPaymentRepositoryPG(pgHost, pgPort, pgUser, pgPassword, pgDB, pgPaymentTable)

	if err != nil {
		panic(err)
	}

	userRepository, err = repositories.NewUserRepositoryPG(pgHost, pgPort, pgUser, pgPassword, pgDB, pgUserTable, pgTokenTable)

	if err != nil {
		panic(err)
	}

	http.HandleFunc("/api/v1/merchant/pay", payHandler)

	_ = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		createPay(w, r)
		break
	case http.MethodGet:
		getPay(w, r)
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func createPay(w http.ResponseWriter, r *http.Request) {
	userPayToken := r.FormValue("user_pay_token")
	amount, err := strconv.Atoi(r.FormValue("amount"))

	if userPayToken == "" || err != nil || amount <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	claim := jwtManager.GetTokenClaim(userPayToken)

	if claim == nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	//TODO перекинуть получение в интерфейс
	user := userRepository.GetUserById(int((*claim)["id"].(float64)))
	token := userRepository.GetTokenById(user.TokenId)

	paymentId := paymentRepository.Create(float64(amount), user.Id, "", (*claim)["key"].(string))

	resp := qiwiClient.CreatePayment(qiwiSdkModels.CreatePayment{
		Amount: qiwiSdkSubModels.Amount{
			Currency: "RUB",
			Value:    fmt.Sprintf("%.2f", float64(amount)),
		},
		PaymentMethod: qiwiSdkSubModels.PaymentMethod{
			Type:         "TOKEN",
			PaymentToken: token.Token,
		},
		Customer: qiwiSdkSubModels.Customer{
			Account: strconv.Itoa(user.Id),
			Phone:   user.Phone,
		},
		Flags: []string{"BIND_PAYMENT_TOKEN"},
	}, strconv.Itoa(paymentId))

	paymentRepository.UpdatePayment(paymentId, resp.Status.Value, resp.BillId)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     paymentId,
		"amount": float64(amount),
		"status": resp.Status.Value,
	})
}

func getPay(w http.ResponseWriter, r *http.Request) {
	paymentId, err := strconv.Atoi(r.FormValue("payment_id"))

	if err != nil || paymentId <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	payment := paymentRepository.GetById(paymentId)

	resp := qiwiClient.GetPayment(strconv.Itoa(paymentId))

	//TODO сделать обновления статусу через вебхук
	if payment.Status != resp.Status.Value {
		paymentRepository.UpdateStatus(paymentId, resp.Status.Value)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":     paymentId,
		"amount": payment.Amount,
		"status": resp.Status.Value,
	})
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"

	"rabbitmq-go-lab/internal/config"
	"rabbitmq-go-lab/internal/rabbit"
)

func main() {
	_ = godotenv.Load()
	c := config.Load()

	conn, ch, err := rabbit.Dial(c.URL)
	if err != nil { log.Fatalf("dial: %v", err) }
	defer conn.Close(); defer ch.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<h1>RabbitMQ Monitor UI (Demo)</h1>
		<form method="POST" action="/publish">
			Message: <input name="msg" value="hello" />
			<label><input type="checkbox" name="fail" /> force fail</label>
			<button type="submit">Publish</button>
		</form>
		<p>Use Grafana: <code>http://localhost:3000</code> (admin/admin)</p>
		<p>Rabbit Management: <code>http://localhost:15672</code> (guest/guest)</p>
		`)
	})

	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		fail := r.Form.Get("fail") == "on"
		body := map[string]any{"message": r.Form.Get("msg"), "ts": time.Now().Format(time.RFC3339Nano)}
		b, _ := json.Marshal(body)
		headers := amqp.Table{"schemaVersion": 1}
		if fail { headers["forceFail"] = true }
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		err := rabbit.PublishConfirm(ctx, ch, c.Exchange, "demo.info", amqp.Publishing{
			ContentType: "application/json",
			DeliveryMode: amqp.Persistent,
			Body: b,
			Headers: headers,
			MessageId: fmt.Sprintf("ui-%d", time.Now().UnixNano()),
			Type: "demo",
		})
		cancel()
		if err != nil {
			w.WriteHeader(500)
			_, _ = w.Write([]byte("publish error: "+err.Error()))
			return
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	log.Println("monitor-ui on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

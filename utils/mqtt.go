package utils

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var (
	mqttClient mqtt.Client
)

// MQTTConfig holds the MQTT broker configuration
type MQTTConfig struct {
	Broker   string
	Port     int
	Username string
	Password string
	ClientID string
	Topic    string
}

// GetMQTTConfig returns the MQTT configuration from environment variables
func GetMQTTConfig() MQTTConfig {
	return MQTTConfig{
		Broker:   os.Getenv("MQTT_BROKER"),
		Port:     1883,
		Username: os.Getenv("MQTT_USERNAME"),
		Password: os.Getenv("MQTT_PASSWORD"),
		ClientID: "lights_server",
		Topic:    "esp32/lights",
	}
}

// ConnectMQTT connects to the MQTT broker
func ConnectMQTT() error {
	config := GetMQTTConfig()
	if config.Broker == "" || config.Username == "" || config.Password == "" {
		return fmt.Errorf("MQTT configuration is incomplete")
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s:%d", config.Broker, config.Port))
	opts.SetUsername(config.Username)
	opts.SetPassword(config.Password)
	opts.SetClientID(config.ClientID)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		fmt.Printf("Connection lost: %v\n", err)
	})
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		fmt.Println("Connected to MQTT broker")
	})

	mqttClient = mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %v", token.Error())
	}

	return nil
}

// PublishRGB publishes RGB values to the MQTT topic
func PublishRGB(rgb RGB) error {
	if mqttClient == nil || !mqttClient.IsConnected() {
		return fmt.Errorf("MQTT client is not connected")
	}

	message := fmt.Sprintf("{%d,%d,%d}", rgb.R, rgb.G, rgb.B)
	token := mqttClient.Publish(GetMQTTConfig().Topic, 0, false, message)
	token.WaitTimeout(5 * time.Second)
	
	if token.Error() != nil {
		return fmt.Errorf("failed to publish message: %v", token.Error())
	}

	return nil
} 
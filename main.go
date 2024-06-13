package main

import (
	"database/sql"
	"fmt"

	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // Assuming MySQL driver for database interaction
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"tinygo.org/x/bluetooth"
)

var (
	adapter           = bluetooth.DefaultAdapter
	PixelsServiceUUID = bluetooth.NewUUID([16]byte{0x6e, 0x40, 0x00, 0x01, 0xb5, 0xa3, 0xf3, 0x93, 0xe0, 0xa9, 0xe5, 0x0e, 0x24, 0xdc, 0xca, 0x9e})
	PixelsNotifyUUID  = bluetooth.NewUUID([16]byte{0x6e, 0x40, 0x00, 0x01, 0xb5, 0xa3, 0xf3, 0x93, 0xe0, 0xa9, 0xe5, 0x0e, 0x24, 0xdc, 0xca, 0x9e})
	PixelsWriteUUID   = bluetooth.NewUUID([16]byte{0x6e, 0x40, 0x00, 0x02, 0xb5, 0xa3, 0xf3, 0x93, 0xe0, 0xa9, 0xe5, 0x0e, 0x24, 0xdc, 0xca, 0x9e})
	db                *sql.DB
)

type MessageType byte

const (
	RollStateID MessageType = 3
)

type RollStateMessage struct {
	DeviceID       string
	RollState      uint8
	CurrentFace    uint8
	BatteryLevel   uint8
	IsCharging     bool
	LEDCount       uint8
	DesignAndColor uint8
	AdditionalInfo string // Any additional info you might need
}

const (
	dsn           = "root:password123@tcp(127.0.0.1:3306)/dicemaster" // Replace with your database connection string
	migrationsDir = "migrations"                                      // Directory where your migration files are stored
)

func main() {
	var err error
	// Initialize database connection
	db, err = sql.Open("mysql", "root:password123@tcp(127.0.0.1:3306)/dicemaster")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := runMigrations(db, migrationsDir); err != nil {
		log.Fatalf("Migration failed: %s", err)
	}

	// Enable BLE interface
	err = adapter.Enable()
	if err != nil {
		log.Fatal("Failed to enable BLE stack:", err)
	}

	// Add a delay to ensure the BLE stack has time to initialize
	time.Sleep(5 * time.Second)

	// Start scanning
	println("scanning...")
	err = adapter.Scan(scanHandler)
	if err != nil {
		log.Fatal("Failed to start scan:", err)
	}

	// Keep the program running to allow asynchronous operations to complete
	select {}
}

func scanHandler(adapter *bluetooth.Adapter, device bluetooth.ScanResult) {
	println("found device:", device.Address.String(), device.RSSI, device.LocalName())

	if device.LocalName() == "Galaxy" {
		fmt.Printf("Found target device: %s [%s]\n", device.LocalName(), device.Address.String())

		// Stop scanning once we found the target device
		err := adapter.StopScan()
		if err != nil {
			fmt.Printf("Failed to stop scan: %s\n", err)
			return
		}

		// Connect to the device
		go connectToDevice(device.Address) // Use a goroutine to prevent blocking
	}
}

func connectToDevice(address bluetooth.Address) {
	fmt.Printf("Connecting to device: %s\n", address.String())
	device, err := adapter.Connect(address, bluetooth.ConnectionParams{})
	if err != nil {
		fmt.Printf("Failed to connect to device: %s\n", err)
		return
	}
	fmt.Printf("Connected to device: %s\n", address.String())

	// Discover services
	srvcs, err := device.DiscoverServices(nil)
	if err != nil {
		fmt.Printf("Failed to discover services: %s\n", err)
		return
	}

	for _, srvc := range srvcs {
		fmt.Printf("Service UUID: %s\n", srvc.UUID().String())
		chars, err := srvc.DiscoverCharacteristics(nil)
		if err != nil {
			fmt.Printf("Failed to discover characteristics: %s\n", err)
			return
		}

		for _, char := range chars {
			fmt.Printf("Characteristic UUID: %s\n", char.UUID().String())

			// Subscribe to notifications if the characteristic matches PixelsNotifyUUID
			if char.UUID().String() == (PixelsNotifyUUID.String()) {
				go subscribeToNotifications(&char, address.String()) // Pass device ID to use in database insert
			}
		}
	}
}

func subscribeToNotifications(char *bluetooth.DeviceCharacteristic, deviceID string) {
	fmt.Println("Subscribing to notifications...")
	err := char.EnableNotifications(func(buf []byte) {
		fmt.Printf("Received data: %x\n", buf)
		if len(buf) > 0 {
			message, err := deserializeMessage(buf, deviceID)
			if err != nil {
				fmt.Printf("Failed to deserialize message: %s\n", err)
				return
			}
			handleMessage(message)
		} else {
			fmt.Println("Received empty notification")
		}
	})
	if err != nil {
		fmt.Printf("Failed to subscribe to notifications: %s\n", err)
	} else {
		fmt.Println("Successfully subscribed to notifications")
	}
}

func deserializeMessage(data []byte, deviceID string) (RollStateMessage, error) {
	if len(data) < 1 {
		return RollStateMessage{}, fmt.Errorf("can't deserialize an empty buffer")
	}

	msgType := MessageType(data[0])

	if msgType != RollStateID {
		return RollStateMessage{}, fmt.Errorf("unknown message type: %d", msgType)
	}

	if len(data) < 3 {
		return RollStateMessage{}, fmt.Errorf("RollState message too short")
	}

	// Here you would parse other information like battery level, LED count, etc., from the notification data
	batteryLevel := uint8(0)   // Placeholder value, parse from `data`
	isCharging := false        // Placeholder value, parse from `data`
	ledCount := uint8(0)       // Placeholder value, parse from `data`
	designAndColor := uint8(0) // Placeholder value, parse from `data`

	return RollStateMessage{
		DeviceID:       deviceID,
		RollState:      data[1],
		CurrentFace:    data[2],
		BatteryLevel:   batteryLevel,
		IsCharging:     isCharging,
		LEDCount:       ledCount,
		DesignAndColor: designAndColor,
		AdditionalInfo: "Sample additional info", // Add any additional info you need
	}, nil
}

func handleMessage(message RollStateMessage) {
	fmt.Printf("RollState message received:\n")
	fmt.Printf("Device ID: %s\n", message.DeviceID)
	fmt.Printf("Roll State: %d\n", message.RollState)
	fmt.Printf("Current Face: %d\n", message.CurrentFace)
	fmt.Printf("Battery Level: %d\n", message.BatteryLevel)
	fmt.Printf("Is Charging: %v\n", message.IsCharging)
	fmt.Printf("LED Count: %d\n", message.LEDCount)
	fmt.Printf("Design and Color: %d\n", message.DesignAndColor)
	fmt.Printf("Additional Info: %s\n", message.AdditionalInfo)

	// Insert data into database
	query := `INSERT INTO roll_state_data (device_id, roll_state, current_face, timestamp, battery_level, is_charging, led_count, design_and_color, additional_info) VALUES (?, ?, ?, NOW(), ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, message.DeviceID, message.RollState, message.CurrentFace, message.BatteryLevel, message.IsCharging, message.LEDCount, message.DesignAndColor, message.AdditionalInfo)
	if err != nil {
		fmt.Printf("Failed to insert data into database: %s\n", err)
	} else {
		fmt.Println("Data successfully inserted into database")
	}
}

func runMigrations(db *sql.DB, migrationsDir string) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create MySQL driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsDir,
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	fmt.Println("Migrations applied successfully!")
	return nil
}

func initializeDatabase() (*sql.DB, error) {
	dsn := "root:password123@tcp(127.0.0.1:3306)/dicemaster?multiStatements=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not create MySQL driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("could not run up migrations: %w", err)
	}

	return db, nil
}

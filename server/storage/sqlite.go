package storage

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/accnotify/server/model"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements storage using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// initTables creates the necessary tables
func (s *SQLiteStorage) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS devices (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_key TEXT UNIQUE NOT NULL,
			public_key TEXT,
			name TEXT,
<<<<<<< HEAD
			device_token TEXT,
			platform TEXT,
=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_seen DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			device_id INTEGER NOT NULL,
			message_id TEXT UNIQUE NOT NULL,
			title TEXT,
			body TEXT,
			group_name TEXT,
			icon TEXT,
			url TEXT,
			sound TEXT,
			badge INTEGER DEFAULT 0,
			encrypted_payload BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			delivered BOOLEAN DEFAULT FALSE,
			FOREIGN KEY (device_id) REFERENCES devices(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_device_id ON messages(device_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Device operations

// GetDeviceByKey retrieves a device by its key
func (s *SQLiteStorage) GetDeviceByKey(deviceKey string) (*model.Device, error) {
	device := &model.Device{}
	err := s.db.QueryRow(
<<<<<<< HEAD
		`SELECT id, device_key, public_key, name, device_token, platform, created_at, last_seen 
		 FROM devices WHERE device_key = ?`,
		deviceKey,
	).Scan(&device.ID, &device.DeviceKey, &device.PublicKey, &device.Name, &device.DeviceToken, &device.Platform, &device.CreatedAt, &device.LastSeen)
=======
		`SELECT id, device_key, public_key, name, created_at, last_seen 
		 FROM devices WHERE device_key = ?`,
		deviceKey,
	).Scan(&device.ID, &device.DeviceKey, &device.PublicKey, &device.Name, &device.CreatedAt, &device.LastSeen)
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return device, nil
}

// CreateDevice creates a new device
func (s *SQLiteStorage) CreateDevice(device *model.Device) error {
	result, err := s.db.Exec(
<<<<<<< HEAD
		`INSERT INTO devices (device_key, public_key, name, device_token, platform, created_at, last_seen) 
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		device.DeviceKey, device.PublicKey, device.Name, device.DeviceToken, device.Platform, time.Now(), time.Now(),
=======
		`INSERT INTO devices (device_key, public_key, name, created_at, last_seen) 
		 VALUES (?, ?, ?, ?, ?)`,
		device.DeviceKey, device.PublicKey, device.Name, time.Now(), time.Now(),
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	device.ID = id
	return nil
}

// UpdateDevicePublicKey updates the public key for a device
func (s *SQLiteStorage) UpdateDevicePublicKey(deviceKey, publicKey string) error {
	_, err := s.db.Exec(
		`UPDATE devices SET public_key = ?, last_seen = ? WHERE device_key = ?`,
		publicKey, time.Now(), deviceKey,
	)
	return err
}

<<<<<<< HEAD
// UpdateDeviceToken updates the device token for APNs
func (s *SQLiteStorage) UpdateDeviceToken(deviceKey, deviceToken string) error {
	_, err := s.db.Exec(
		`UPDATE devices SET device_token = ?, last_seen = ? WHERE device_key = ?`,
		deviceToken, time.Now(), deviceKey,
	)
	return err
}

// UpdateDevicePlatform updates the platform for a device
func (s *SQLiteStorage) UpdateDevicePlatform(deviceKey, platform string) error {
	_, err := s.db.Exec(
		`UPDATE devices SET platform = ?, last_seen = ? WHERE device_key = ?`,
		platform, time.Now(), deviceKey,
	)
	return err
}

// UpdateDeviceName updates the name for a device
func (s *SQLiteStorage) UpdateDeviceName(deviceKey, name string) error {
	_, err := s.db.Exec(
		`UPDATE devices SET name = ?, last_seen = ? WHERE device_key = ?`,
		name, time.Now(), deviceKey,
	)
	return err
}

=======
>>>>>>> 287a33766748e36d05bf23587ea969d835085abe
// UpdateDeviceLastSeen updates the last seen timestamp
func (s *SQLiteStorage) UpdateDeviceLastSeen(deviceKey string) error {
	_, err := s.db.Exec(
		`UPDATE devices SET last_seen = ? WHERE device_key = ?`,
		time.Now(), deviceKey,
	)
	return err
}

// Message operations

// CreateMessage stores a new message
func (s *SQLiteStorage) CreateMessage(msg *model.Message) error {
	result, err := s.db.Exec(
		`INSERT INTO messages (device_id, message_id, title, body, group_name, icon, url, sound, badge, encrypted_payload, created_at, delivered) 
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		msg.DeviceID, msg.MessageID, msg.Title, msg.Body, msg.Group, msg.Icon, msg.URL, msg.Sound, msg.Badge, msg.EncryptedPayload, time.Now(), false,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	msg.ID = id
	return nil
}

// MarkMessageDelivered marks a message as delivered
func (s *SQLiteStorage) MarkMessageDelivered(messageID string) error {
	_, err := s.db.Exec(
		`UPDATE messages SET delivered = TRUE WHERE message_id = ?`,
		messageID,
	)
	return err
}

// GetUndeliveredMessages retrieves undelivered messages for a device
func (s *SQLiteStorage) GetUndeliveredMessages(deviceID int64) ([]*model.Message, error) {
	rows, err := s.db.Query(
		`SELECT id, device_id, message_id, title, body, group_name, icon, url, sound, badge, encrypted_payload, created_at, delivered 
		 FROM messages 
		 WHERE device_id = ? AND delivered = FALSE 
		 ORDER BY created_at ASC`,
		deviceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.DeviceID, &msg.MessageID, &msg.Title, &msg.Body,
			&msg.Group, &msg.Icon, &msg.URL, &msg.Sound, &msg.Badge,
			&msg.EncryptedPayload, &msg.CreatedAt, &msg.Delivered,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMessageHistory retrieves message history for a device
func (s *SQLiteStorage) GetMessageHistory(deviceID int64, limit, offset int) ([]*model.Message, error) {
	rows, err := s.db.Query(
		`SELECT id, device_id, message_id, title, body, group_name, icon, url, sound, badge, created_at, delivered 
		 FROM messages 
		 WHERE device_id = ? 
		 ORDER BY created_at DESC 
		 LIMIT ? OFFSET ?`,
		deviceID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*model.Message
	for rows.Next() {
		msg := &model.Message{}
		err := rows.Scan(
			&msg.ID, &msg.DeviceID, &msg.MessageID, &msg.Title, &msg.Body,
			&msg.Group, &msg.Icon, &msg.URL, &msg.Sound, &msg.Badge,
			&msg.CreatedAt, &msg.Delivered,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// DeleteOldMessages deletes messages older than the specified duration
func (s *SQLiteStorage) DeleteOldMessages(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result, err := s.db.Exec(
		`DELETE FROM messages WHERE created_at < ?`,
		cutoff,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

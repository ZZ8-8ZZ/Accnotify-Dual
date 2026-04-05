# APNs Configuration

## Environment Variables

To enable APNs push notifications for iOS devices (Bark compatibility), set the following environment variables:

### Required Variables

```bash
# Enable APNs service
ACCNOTIFY_APNS_ENABLED=true

# APNs Topic (your app's bundle identifier)
ACCNOTIFY_APNS_TOPIC=com.yourcompany.yourapp
```

### Authentication Methods

Choose one of the following authentication methods:

#### Method 1: Token-based Authentication (Recommended)

```bash
# APNs Key ID (from Apple Developer portal)
ACCNOTIFY_APNS_KEY_ID=ABC123XYZ

# APNs Team ID (from Apple Developer portal)
ACCNOTIFY_APNS_TEAM_ID=DEF456UVW

# Path to APNs private key file (.p8 format)
ACCNOTIFY_APNS_KEY_FILE=/path/to/APNsAuthKey_ABC123XYZ.p8
```

#### Method 2: Certificate-based Authentication

```bash
# Path to APNs certificate file (.p12 format)
ACCNOTIFY_APNS_CERT_FILE=/path/to/apns_certificate.p12
```

### Optional Variables

```bash
# Use development APNs server (default: false)
ACCNOTIFY_APNS_DEVELOPMENT=false
```

## Setup Instructions

### 1. Create APNs Key (Token-based Auth)

1. Go to [Apple Developer Portal](https://developer.apple.com/account/)
2. Navigate to "Certificates, Identifiers & Profiles"
3. Select "Keys" from the left sidebar
4. Click "+" to create a new key
5. Enter a key name and select "Apple Push Notifications service (APNs)"
6. Download the `.p8` key file (you can only download it once!)
7. Note down the Key ID and your Team ID

### 2. Create APNs Certificate (Certificate-based Auth)

1. Go to [Apple Developer Portal](https://developer.apple.com/account/)
2. Navigate to "Certificates, Identifiers & Profiles"
3. Select "Certificates" from the left sidebar
4. Click "+" to create a new certificate
5. Choose "Apple Push Notification service SSL (Sandbox & Production)"
6. Select your App ID
7. Follow the instructions to create a CSR (Certificate Signing Request)
8. Download the `.cer` file
9. Convert to `.p12` format:
   ```bash
   # On macOS
   security import certificate.cer -k ~/Library/Keychains/login.keychain-db
   security export certificate -p12 -out apns_certificate.p12
   ```

### 3. Configure Your App Bundle ID

1. Go to [Apple Developer Portal](https://developer.apple.com/account/)
2. Navigate to "Identifiers"
3. Select your App ID
4. Enable "Push Notifications" capability
5. Use the bundle identifier as `ACCNOTIFY_APNS_TOPIC`

### 4. Deploy Configuration

#### Docker Compose

```yaml
version: '3.8'

services:
  accnotify:
    image: accnotify/server:latest
    ports:
      - "8080:8080"
    environment:
      - ACCNOTIFY_HOST=0.0.0.0
      - ACCNOTIFY_PORT=8080
      - ACCNOTIFY_APNS_ENABLED=true
      - ACCNOTIFY_APNS_TOPIC=com.yourcompany.yourapp
      - ACCNOTIFY_APNS_KEY_ID=ABC123XYZ
      - ACCNOTIFY_APNS_TEAM_ID=DEF456UVW
      - ACCNOTIFY_APNS_KEY_FILE=/secrets/apns_key.p8
      - ACCNOTIFY_APNS_DEVELOPMENT=false
    volumes:
      - ./apns_key.p8:/secrets/apns_key.p8:ro
      - ./data:/data
```

#### Docker Run

```bash
docker run -d \
  --name accnotify \
  -p 8080:8080 \
  -e ACCNOTIFY_APNS_ENABLED=true \
  -e ACCNOTIFY_APNS_TOPIC=com.yourcompany.yourapp \
  -e ACCNOTIFY_APNS_KEY_ID=ABC123XYZ \
  -e ACCNOTIFY_APNS_TEAM_ID=DEF456UVW \
  -e ACCNOTIFY_APNS_KEY_FILE=/secrets/apns_key.p8 \
  -v $(pwd)/apns_key.p8:/secrets/apns_key.p8:ro \
  -v $(pwd)/data:/data \
  accnotify/server:latest
```

#### Direct Execution

```bash
export ACCNOTIFY_APNS_ENABLED=true
export ACCNOTIFY_APNS_TOPIC=com.yourcompany.yourapp
export ACCNOTIFY_APNS_KEY_ID=ABC123XYZ
export ACCNOTIFY_APNS_TEAM_ID=DEF456UVW
export ACCNOTIFY_APNS_KEY_FILE=/path/to/apns_key.p8
export ACCNOTIFY_APNS_DEVELOPMENT=false

./server
```

## Testing APNs

### Test with Bark Client

1. Install Bark on your iOS device
2. Open Bark and note your device key
3. Register the device with Accnotify:

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "device_key": "your_bark_device_key",
    "devicetoken": "your_ios_device_token",
    "platform": "ios",
    "name": "My iPhone"
  }'
```

4. Send a test push:

```bash
curl http://localhost:8080/push/your_bark_device_key/Hello/World
```

### Test with Accnotify Android Client

The Android client will continue to use WebSocket connections and is not affected by APNs configuration.

## Troubleshooting

### Common Issues

1. **"APNs configuration validation failed"**
   - Check that all required environment variables are set
   - Verify that key/certificate files exist and are readable
   - Ensure the topic matches your app's bundle identifier

2. **"Failed to send APNs push"**
   - Check device token is valid
   - Verify the device platform is set to "ios"
   - Check APNs service status in Apple Developer Portal
   - Review server logs for detailed error messages

3. **"APNs topic is not configured"**
   - Ensure `ACCNOTIFY_APNS_TOPIC` is set to your app's bundle identifier

4. **Push not received on iOS device**
   - Verify device token is correct
   - Check that notifications are enabled in iOS settings
   - Ensure the device is not in Do Not Disturb mode
   - Try development mode if testing with development build

### Debug Logging

Enable debug logging in the server to see detailed APNs communication:

```bash
# The server logs will show:
# - APNs service initialization status
# - Push notification attempts
# - APNs response codes and reasons
# - Any errors encountered
```

## Security Considerations

1. **Never commit APNs keys or certificates to version control**
2. **Use environment variables or secret management tools**
3. **Restrict file permissions on key/certificate files**
4. **Rotate APNs keys periodically (recommended annually)**
5. **Use separate keys for development and production**
6. **Monitor APNs usage and quotas**

## Production vs Development

### Development Mode
```bash
ACCNOTIFY_APNS_DEVELOPMENT=true
```
- Uses APNs sandbox server
- For testing with development builds
- Device tokens are different from production

### Production Mode
```bash
ACCNOTIFY_APNS_DEVELOPMENT=false
```
- Uses APNs production server
- For App Store/TestFlight builds
- Production device tokens required

## Additional Resources

- [Apple Push Notification Service Documentation](https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server)
- [APNs Authentication Keys](https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/sending_notification_requests_to_apns)
- [APNs Provider API](https://developer.apple.com/documentation/usernotifications/setting_up_a_remote_notification_server/generating_a_remote_notification)
- [Bark GitHub Repository](https://github.com/Finb/Bark)
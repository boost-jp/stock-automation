# Slack Notification Feature

## Overview

The Slack notification feature provides real-time alerts and daily reports to keep users informed about their portfolio status and important market events.

## Features

### 1. Daily Portfolio Reports
- Automatically sends comprehensive portfolio summaries to Slack
- Includes total value, gains/losses, and individual stock performance
- Enhanced formatting with color-coded indicators
- Scheduled to run daily at configured times

### 2. Stock Price Alerts
- Real-time notifications when stocks hit target prices
- Buy/sell signal alerts with current and target price information
- Color-coded alerts (green for buy, red for sell)

### 3. Error Notifications
- Critical system error alerts
- API failure notifications
- Database error alerts
- Panic recovery notifications

### 4. Retry Mechanism
- Automatic retry on network failures (up to 3 attempts)
- Exponential backoff between retries
- Detailed logging of retry attempts

### 5. Transmission Logging
- All notifications are logged to the database
- Tracks notification status (pending, sent, failed)
- Stores metadata for analytics
- Helps with debugging and monitoring

## Configuration

Add the following environment variables to your configuration:

```yaml
slack:
  webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
  channel: "#stock-alerts"  # Optional
  username: "Stock Bot"     # Optional
```

## Usage

### CLI Commands

```bash
# Send daily report immediately
./stock-automation report

# Run scheduler (includes daily reports)
./stock-automation scheduler
```

### Integration Points

The Slack notification service is integrated into:
- Daily portfolio report generation
- Real-time stock monitoring
- Error handling middleware
- Data collection processes

## Database Schema

The notification logs are stored in the `notification_logs` table:

```sql
CREATE TABLE notification_logs (
    id SERIAL PRIMARY KEY,
    notification_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    message TEXT,
    metadata JSONB,
    error_message TEXT,
    attempts INT DEFAULT 1,
    sent_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Testing

Run the notification tests:

```bash
go test -v ./app/infrastructure/notification/
```

## Monitoring

Monitor notification delivery:
- Check `notification_logs` table for delivery status
- Review application logs for retry attempts
- Set up alerts for failed notifications

## Best Practices

1. **Rate Limiting**: The system respects Slack's rate limits
2. **Error Handling**: All errors are logged with context
3. **Retry Logic**: Transient failures are automatically retried
4. **Monitoring**: Regular checks of notification logs ensure delivery

## Troubleshooting

### Common Issues

1. **Webhook URL not configured**
   - Check SLACK_WEBHOOK_URL environment variable
   - Verify webhook URL is valid

2. **Network errors**
   - Check internet connectivity
   - Verify firewall rules allow HTTPS to Slack

3. **Message formatting errors**
   - Check JSON structure in logs
   - Verify special characters are properly escaped

4. **Rate limiting**
   - Monitor notification frequency
   - Implement additional rate limiting if needed
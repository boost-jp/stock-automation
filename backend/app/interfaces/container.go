package interfaces

import (
	"github.com/boost-jp/stock-automation/app/domain"
	"github.com/boost-jp/stock-automation/app/infrastructure/client"
	"github.com/boost-jp/stock-automation/app/infrastructure/config"
	"github.com/boost-jp/stock-automation/app/infrastructure/database"
	"github.com/boost-jp/stock-automation/app/infrastructure/notification"
	"github.com/boost-jp/stock-automation/app/infrastructure/repository"
	"github.com/boost-jp/stock-automation/app/usecase"
)

// Container holds all the dependencies for the application
type Container struct {
	// Infrastructure
	config                    *config.Config
	connectionManager         database.ConnectionManager
	transactionManager        repository.TransactionManager
	stockRepository           repository.StockRepository
	portfolioRepository       repository.PortfolioRepository
	notificationLogRepository repository.NotificationLogRepository
	schedulerLogRepository    repository.SchedulerLogRepository
	stockDataClient           client.StockDataClient
	notificationService       notification.NotificationService

	// Domain Services
	portfolioService         *domain.PortfolioService
	technicalAnalysisService *domain.TechnicalAnalysisService

	// Use Cases
	collectDataUseCase       *usecase.CollectDataUseCase
	portfolioReportUseCase   *usecase.PortfolioReportUseCase
	technicalAnalysisUseCase *usecase.TechnicalAnalysisUseCase

	// Interface
	scheduler *DataScheduler
}

// NewContainer creates a new dependency injection container
func NewContainer(cfg *config.Config) (*Container, error) {
	container := &Container{
		config: cfg,
	}

	// Initialize infrastructure layer
	if err := container.initializeInfrastructure(); err != nil {
		return nil, err
	}

	// Initialize domain layer
	container.initializeDomain()

	// Initialize use case layer
	container.initializeUseCases()

	// Initialize interface layer
	container.initializeInterfaces()

	return container, nil
}

// initializeInfrastructure sets up the infrastructure layer dependencies
func (c *Container) initializeInfrastructure() error {
	// Database connection
	dbConfig := database.DatabaseConfig{
		Host:         c.config.Database.Host,
		Port:         c.config.Database.Port,
		User:         c.config.Database.User,
		Password:     c.config.Database.Password,
		DatabaseName: c.config.Database.DatabaseName,
		MaxOpenConns: 25,
		MaxIdleConns: 5,
		MaxLifetime:  5 * 60, // 5 minutes
	}

	connMgr, err := database.NewConnectionManager(dbConfig)
	if err != nil {
		return err
	}
	c.connectionManager = connMgr

	// Transaction manager
	c.transactionManager = repository.NewTransactionManager(connMgr.GetDB())

	// Repositories
	c.stockRepository = repository.NewStockRepository(connMgr.GetExecutor())
	c.portfolioRepository = repository.NewPortfolioRepository(connMgr.GetExecutor())
	c.notificationLogRepository = repository.NewNotificationLogRepository(connMgr.GetExecutor())
	c.schedulerLogRepository = repository.NewSchedulerLogRepository(connMgr.GetExecutor())

	// External clients
	yahooConfig := client.YahooFinanceConfig{
		BaseURL:       c.config.Yahoo.BaseURL,
		Timeout:       c.config.Yahoo.Timeout,
		RetryCount:    c.config.Yahoo.RetryCount,
		RetryWaitTime: c.config.Yahoo.RetryWaitTime,
		RetryMaxWait:  c.config.Yahoo.RetryMaxWait,
		UserAgent:     c.config.Yahoo.UserAgent,
		RateLimitRPS:  c.config.Yahoo.RateLimitRPS,
	}
	c.stockDataClient = client.NewYahooFinanceClientWithConfig(yahooConfig)

	// Notification service
	slackNotifier := notification.NewSlackNotificationService(
		c.config.Slack.WebhookURL,
		c.config.Slack.Channel,
		c.config.Slack.Username,
	)
	// Set notification log repository if it's a SlackNotifier
	if sn, ok := slackNotifier.(*notification.SlackNotifier); ok {
		sn.SetLogRepository(c.notificationLogRepository)
	}
	c.notificationService = slackNotifier

	return nil
}

// initializeDomain sets up the domain layer services
func (c *Container) initializeDomain() {
	c.portfolioService = domain.NewPortfolioService()
	c.technicalAnalysisService = domain.NewTechnicalAnalysisService()
}

// initializeUseCases sets up the use case layer
func (c *Container) initializeUseCases() {
	c.collectDataUseCase = usecase.NewCollectDataUseCase(
		c.stockRepository,
		c.portfolioRepository,
		c.stockDataClient,
	)

	c.portfolioReportUseCase = usecase.NewPortfolioReportUseCase(
		c.stockRepository,
		c.portfolioRepository,
		c.stockDataClient,
		c.notificationService,
	)

	c.technicalAnalysisUseCase = usecase.NewTechnicalAnalysisUseCase(
		c.stockRepository,
		c.stockDataClient,
	)
}

// initializeInterfaces sets up the interface layer
func (c *Container) initializeInterfaces() {
	c.scheduler = NewDataScheduler(
		c.collectDataUseCase,
		c.portfolioReportUseCase,
	)
	c.scheduler.SetLogRepository(c.schedulerLogRepository)
}

// GetConnectionManager returns the database connection manager
func (c *Container) GetConnectionManager() database.ConnectionManager {
	return c.connectionManager
}

// GetStockRepository returns the stock repository
func (c *Container) GetStockRepository() repository.StockRepository {
	return c.stockRepository
}

// GetPortfolioRepository returns the portfolio repository
func (c *Container) GetPortfolioRepository() repository.PortfolioRepository {
	return c.portfolioRepository
}

// GetCollectDataUseCase returns the collect data use case
func (c *Container) GetCollectDataUseCase() *usecase.CollectDataUseCase {
	return c.collectDataUseCase
}

// GetPortfolioReportUseCase returns the portfolio report use case
func (c *Container) GetPortfolioReportUseCase() *usecase.PortfolioReportUseCase {
	return c.portfolioReportUseCase
}

// GetTechnicalAnalysisUseCase returns the technical analysis use case
func (c *Container) GetTechnicalAnalysisUseCase() *usecase.TechnicalAnalysisUseCase {
	return c.technicalAnalysisUseCase
}

// GetScheduler returns the data scheduler
func (c *Container) GetScheduler() *DataScheduler {
	return c.scheduler
}

// Close cleans up all resources
func (c *Container) Close() error {
	if c.scheduler != nil {
		c.scheduler.Stop()
	}

	if c.connectionManager != nil {
		return c.connectionManager.Close()
	}

	return nil
}

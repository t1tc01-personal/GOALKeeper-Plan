# Deployment Guide: GOALKeeper Production Deployment

**Last Updated**: January 27, 2026
**For**: DevOps engineers and production deployment
**Audience**: Platform, infrastructure teams

---

## Table of Contents

1. [Deployment Options](#deployment-options)
2. [Environment Configuration](#environment-configuration)
3. [Database Deployment](#database-deployment)
4. [Backend Deployment](#backend-deployment)
5. [Frontend Deployment](#frontend-deployment)
6. [Monitoring & Logging](#monitoring--logging)
7. [Scaling Considerations](#scaling-considerations)
8. [Disaster Recovery](#disaster-recovery)
9. [Security Checklist](#security-checklist)

---

## Deployment Options

### Option 1: Vercel + AWS (Recommended for MVP)

**Architecture**:
- **Frontend**: Vercel (zero-config Next.js deployment)
- **Backend**: AWS EC2 or ECS
- **Database**: AWS RDS for PostgreSQL
- **CDN**: Cloudfront (via Vercel)
- **DNS**: Route 53

**Advantages**:
- Fast frontend deployment
- Built-in HTTPS and caching
- Automatic scaling
- Pay-as-you-go pricing

**Estimated Cost**: $10-50/month

### Option 2: DigitalOcean App Platform

**Architecture**:
- All services on DigitalOcean (App Platform)
- Managed PostgreSQL database
- Built-in load balancing
- One-click deployments

**Advantages**:
- Simple unified dashboard
- Lower learning curve
- Predictable pricing ($12/month minimum)

**Estimated Cost**: $12-30/month

### Option 3: Google Cloud Run + Cloud SQL

**Architecture**:
- **Frontend**: Cloud Storage + Cloud CDN
- **Backend**: Cloud Run (serverless)
- **Database**: Cloud SQL for PostgreSQL

**Advantages**:
- Truly serverless backend
- Auto-scaling
- Pay only for what you use

### Option 4: Kubernetes (Self-Managed)

**Architecture**:
- Backend and frontend in Kubernetes pods
- PostgreSQL operator managed database
- Ingress controller for routing

**Advantages**:
- Full control
- Highly scalable
- Enterprise-ready

**Disadvantages**:
- Complex to manage
- Requires DevOps expertise
- Overkill for MVP

---

## Environment Configuration

### Secrets Management

**Never commit secrets**. Use environment-specific configuration:

```bash
# Production .env (NOT committed)
DATABASE_URL=postgres://user:pass@prod-db.rds.amazonaws.com:5432/goalkeeper_prod
BACKEND_PORT=8080
LOG_LEVEL=info
JWT_SECRET=<generate-strong-secret>
CORS_ALLOWED_ORIGINS=https://goalkeeper.example.com
```

### Backend Environment Variables

```
# Database
DATABASE_URL              # PostgreSQL connection string
DB_MAX_CONNECTIONS=25    # Connection pool size

# Server
BACKEND_PORT=8080
LOG_LEVEL=info           # debug, info, warn, error
ENVIRONMENT=production

# Security
JWT_SECRET               # Sign JWT tokens (if using JWT)
CORS_ALLOWED_ORIGINS     # Comma-separated allowed origins
SECURE_COOKIES=true      # HTTPS only

# Monitoring (optional)
SENTRY_DSN              # Error reporting
DATADOG_API_KEY         # Metrics collection
```

### Frontend Environment Variables

```
# In Vercel dashboard or .env.production
NEXT_PUBLIC_API_URL=https://api.goalkeeper.example.com
NEXT_PUBLIC_ENVIRONMENT=production
```

---

## Database Deployment

### AWS RDS for PostgreSQL

#### 1. Create RDS Instance

```bash
aws rds create-db-instance \
  --db-instance-identifier goalkeeper-prod-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 15.3 \
  --master-username admin \
  --master-user-password <SECURE-PASSWORD> \
  --allocated-storage 20 \
  --storage-type gp3 \
  --vpc-security-group-ids sg-xxxxx \
  --db-subnet-group-name default \
  --backup-retention-period 7 \
  --multi-az
```

#### 2. Get Connection String

```bash
# Get endpoint
aws rds describe-db-instances \
  --db-instance-identifier goalkeeper-prod-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text

# Format connection string
DATABASE_URL=postgres://admin:<PASSWORD>@<ENDPOINT>:5432/goalkeeper
```

#### 3. Run Migrations

```bash
# From backend directory with production DATABASE_URL set
atlas migrate apply --dir file://migrations --url $DATABASE_URL
```

#### 4. Backup Configuration

```bash
# Enable automated backups (7 days retention)
# Already configured with --backup-retention-period 7

# Create manual backup
aws rds create-db-snapshot \
  --db-instance-identifier goalkeeper-prod-db \
  --db-snapshot-identifier goalkeeper-prod-db-backup-$(date +%Y%m%d)
```

### PostgreSQL Managed Databases (DigitalOcean)

```bash
# Create via doctl CLI
doctl databases create goalkeeper-prod \
  --engine pg \
  --version 15 \
  --region nyc3 \
  --num-nodes 1

# Get connection string
doctl databases connection goalkeeper-prod
```

---

## Backend Deployment

### Option 1: AWS EC2

#### 1. Launch EC2 Instance

```bash
# Launch Ubuntu 22.04 LTS instance
aws ec2 run-instances \
  --image-id ami-0c55b159cbfafe1f0 \
  --instance-type t3.micro \
  --security-group-ids sg-xxxxx \
  --subnet-id subnet-xxxxx \
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=goalkeeper-backend}]'
```

#### 2. Connect and Setup

```bash
# SSH into instance
ssh -i goalkeeper-key.pem ubuntu@<PUBLIC-IP>

# Install dependencies
sudo apt update
sudo apt install -y golang-go git

# Clone repository
git clone https://github.com/your-org/GOALKeeper-Plan.git
cd GOALKeeper-Plan/backend

# Set environment variables
cat > .env << EOF
DATABASE_URL=postgres://...
LOG_LEVEL=info
BACKEND_PORT=8080
EOF

# Build application
go build -o bin/main ./cmd/main.go

# Run with systemd
sudo tee /etc/systemd/system/goalkeeper.service > /dev/null << EOF
[Unit]
Description=GOALKeeper Backend API
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/home/ubuntu/GOALKeeper-Plan/backend
EnvironmentFile=/home/ubuntu/GOALKeeper-Plan/backend/.env
ExecStart=/home/ubuntu/GOALKeeper-Plan/backend/bin/main
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

# Start service
sudo systemctl enable goalkeeper
sudo systemctl start goalkeeper
sudo systemctl status goalkeeper
```

### Option 2: Docker + ECS (AWS)

#### 1. Build Docker Image

```bash
# In backend directory
docker build -t goalkeeper-api:latest .
```

#### 2. Push to ECR

```bash
# Create ECR repository
aws ecr create-repository --repository-name goalkeeper-api

# Get login token
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin 123456789.dkr.ecr.us-east-1.amazonaws.com

# Tag image
docker tag goalkeeper-api:latest \
  123456789.dkr.ecr.us-east-1.amazonaws.com/goalkeeper-api:latest

# Push image
docker push 123456789.dkr.ecr.us-east-1.amazonaws.com/goalkeeper-api:latest
```

#### 3. Deploy to ECS

```bash
# Create ECS cluster (UI or CLI)
aws ecs create-cluster --cluster-name goalkeeper-prod

# Create task definition
aws ecs register-task-definition \
  --family goalkeeper-api \
  --container-definitions '[{"name":"api","image":"123456789.dkr.ecr.us-east-1.amazonaws.com/goalkeeper-api:latest","portMappings":[{"containerPort":8080,"hostPort":8080}],"environment":[{"name":"DATABASE_URL","value":"postgres://..."}]}]'

# Create service
aws ecs create-service \
  --cluster goalkeeper-prod \
  --service-name goalkeeper-api \
  --task-definition goalkeeper-api \
  --desired-count 2 \
  --launch-type FARGATE
```

### Option 3: Vercel Functions (Serverless)

```bash
# Deploy Go functions to Vercel
# Move Go handler to api/ directory
# Deploy with vercel CLI
vercel deploy
```

---

## Frontend Deployment

### Option 1: Vercel (Recommended)

#### 1. Connect Git Repository

```bash
# https://vercel.com/new
# Select GitHub repository
# Vercel auto-detects Next.js configuration
```

#### 2. Configure Environment Variables

In Vercel dashboard:
```
NEXT_PUBLIC_API_URL = https://api.goalkeeper.example.com
```

#### 3. Deploy

```bash
# Automatic on push to main
# Or manual deploy
vercel deploy --prod
```

### Option 2: AWS S3 + CloudFront

```bash
# Build Next.js for static export
npm run build

# Upload to S3
aws s3 sync out/ s3://goalkeeper-web/ --delete

# Invalidate CloudFront cache
aws cloudfront create-invalidation \
  --distribution-id E123EXAMPLE \
  --paths "/*"
```

### Option 3: DigitalOcean App Platform

```bash
# Using doctl
doctl apps create --spec app.yaml

# Or via web UI
# Connect GitHub repository
# Auto-builds and deploys
```

---

## Domain and DNS Setup

### Set Custom Domain

**AWS Route 53**:
```bash
# Create hosted zone
aws route53 create-hosted-zone \
  --name goalkeeper.example.com \
  --caller-reference $(date +%s)

# Add A record (pointing to load balancer)
aws route53 change-resource-record-sets \
  --hosted-zone-id Z123EXAMPLE \
  --change-batch file://dns-changes.json
```

**dns-changes.json**:
```json
{
  "Changes": [
    {
      "Action": "CREATE",
      "ResourceRecordSet": {
        "Name": "goalkeeper.example.com",
        "Type": "A",
        "AliasTarget": {
          "HostedZoneId": "Z35SXDOTRQ7X7K",
          "DNSName": "d1234567.cloudfront.net",
          "EvaluateTargetHealth": false
        }
      }
    }
  ]
}
```

### HTTPS Certificate

**AWS Certificate Manager** (free):
```bash
aws acm request-certificate \
  --domain-name goalkeeper.example.com \
  --validation-method DNS
```

Verify DNS records and certificate auto-provisions in 5-15 minutes.

---

## Load Balancing

### AWS Application Load Balancer

```bash
# Create ALB
aws elbv2 create-load-balancer \
  --name goalkeeper-alb \
  --subnets subnet-xxxxx subnet-yyyyy \
  --security-groups sg-xxxxx

# Create target group
aws elbv2 create-target-group \
  --name goalkeeper-api-tg \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-xxxxx

# Register targets (EC2 instances)
aws elbv2 register-targets \
  --target-group-arn arn:aws:elasticloadbalancing:... \
  --targets Id=i-xxxxx Id=i-yyyyy

# Create listener
aws elbv2 create-listener \
  --load-balancer-arn arn:aws:elasticloadbalancing:... \
  --protocol HTTP \
  --port 80
```

### Auto-Scaling

```bash
# Create launch template
aws ec2 create-launch-template \
  --launch-template-name goalkeeper-template \
  --launch-template-data file://launch-template.json

# Create Auto Scaling group
aws autoscaling create-auto-scaling-group \
  --auto-scaling-group-name goalkeeper-asg \
  --launch-template LaunchTemplateName=goalkeeper-template \
  --min-size 1 \
  --max-size 5 \
  --desired-capacity 2 \
  --load-balancer-names goalkeeper-alb
```

---

## Monitoring & Logging

### CloudWatch (AWS)

```bash
# Create log group
aws logs create-log-group --log-group-name /goalkeeper/api

# Backend sends logs
# Via STDOUT (containerized) or CloudWatch agent

# View logs
aws logs tail /goalkeeper/api --follow
```

### Application Performance Monitoring

**Option 1: AWS X-Ray**

```go
// In backend code
import "github.com/aws/aws-xray-sdk-go/xray"

// Trace HTTP requests
// Trace database queries
```

**Option 2: Datadog**

```bash
# Add Datadog agent
# Set DATADOG_API_KEY environment variable
# Backend sends metrics and logs
```

### Error Tracking (Sentry)

```bash
# Set SENTRY_DSN environment variable
# Backend captures and reports errors
```

---

## Scaling Considerations

### Database Scaling

**Read Replicas**:
```bash
aws rds create-db-instance-read-replica \
  --db-instance-identifier goalkeeper-prod-db-replica \
  --source-db-instance-identifier goalkeeper-prod-db
```

**Vertical Scaling**:
```bash
aws rds modify-db-instance \
  --db-instance-identifier goalkeeper-prod-db \
  --db-instance-class db.t3.small \
  --apply-immediately
```

### Backend Scaling

- Use load balancer with auto-scaling groups
- Scale up when CPU > 70% or memory > 80%
- Connections per instance: ~25-50 (limited by DB pool)

### Frontend Caching

- Vercel auto-caches static assets
- Set appropriate Cache-Control headers
- Use CDN for media files

---

## Disaster Recovery

### Backup Strategy

- **Database**: Automated backups (7-day retention)
- **Manual backup** before major changes:

```bash
aws rds create-db-snapshot \
  --db-instance-identifier goalkeeper-prod-db \
  --db-snapshot-identifier goalkeeper-prod-db-backup-$(date +%Y%m%d)
```

### Recovery Procedures

**Restore from backup**:

```bash
# Restore to new instance
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier goalkeeper-prod-db-restored \
  --db-snapshot-identifier goalkeeper-prod-db-backup-20260127

# Point DNS to new instance
# Update connection string
```

**Rollback**:

```bash
# Revert to previous deployment
vercel rollback      # Frontend
git revert <commit>  # Backend, then redeploy
```

---

## Security Checklist

### Infrastructure

- [ ] Database in private subnet (no public access)
- [ ] Security groups restrict ingress to necessary ports
- [ ] HTTPS enforced (HTTP redirects to HTTPS)
- [ ] SSL certificate valid and auto-renewing
- [ ] DDoS protection enabled (AWS Shield, Cloudflare)

### Application

- [ ] Environment variables not committed
- [ ] Secrets rotated regularly
- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (parameterized queries)
- [ ] CORS properly configured
- [ ] Rate limiting enabled
- [ ] Logging enabled (no sensitive data)

### Monitoring

- [ ] Error tracking enabled (Sentry)
- [ ] Performance monitoring active (Datadog, X-Ray)
- [ ] Log aggregation configured (CloudWatch, ELK)
- [ ] Alerts set up for critical errors
- [ ] Backup verification tests scheduled

### Access Control

- [ ] SSH key access only (no passwords)
- [ ] AWS IAM roles used (no root account)
- [ ] Developers have limited permissions
- [ ] Deployment restricted to specific users/roles

---

## Cost Optimization

### Development vs Production

| Resource | Dev | Prod |
|----------|-----|------|
| Database | db.t3.micro | db.t3.small |
| EC2 instances | t3.micro | t3.small (×2) |
| Frontend CDN | - | Yes |
| Monitoring | Basic | Full |

### Estimated Monthly Costs

| Service | Quantity | Price/month |
|---------|----------|------------|
| RDS PostgreSQL | 1 | $12-30 |
| EC2 instances | 2 | $15-30 |
| Load balancer | 1 | $15 |
| Data transfer | ~100GB | $5-15 |
| Frontend (Vercel) | - | Free-$25 |
| **Total** | | **$60-120** |

---

## Post-Deployment Validation

```bash
# 1. Health check
curl https://api.goalkeeper.example.com/health

# 2. Create workspace
curl -X POST https://api.goalkeeper.example.com/api/v1/notion/workspaces \
  -H "X-User-ID: $(uuidgen)" \
  -d '{"name": "Production Test"}'

# 3. Monitor error rates (Sentry, CloudWatch)
# Should see 0 errors for new deployments

# 4. Performance check
# Frontend load time < 3s
# API response time < 200ms
```

---

## Rollback Procedure

**If deployment causes issues**:

```bash
# 1. Revert frontend
vercel rollback

# 2. Revert backend
git revert <commit-hash>
git push origin main
# ECS auto-redeploys, or manually:
aws ecs update-service \
  --cluster goalkeeper-prod \
  --service goalkeeper-api \
  --task-definition goalkeeper-api:N
```

---

## Support and Troubleshooting

### Common Issues

**High error rate after deploy**:
1. Check logs: `aws logs tail /goalkeeper/api --follow`
2. Check database connection: `psql $DATABASE_URL`
3. Rollback deployment (see above)

**Slow API responses**:
1. Check database performance: `SELECT * FROM pg_stat_statements`
2. Check slow queries in logs
3. Add indexes if needed

**Frontend 404 errors**:
1. Verify NEXT_PUBLIC_API_URL is correct
2. Check CORS configuration in backend
3. Verify DNS propagation

---

## Next Steps

1. Set up monitoring and alerting
2. Create runbooks for common operations
3. Schedule disaster recovery drills
4. Document team access and procedures
5. Plan horizontal scaling if traffic grows


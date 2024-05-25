# AWS API

Web server that manages AWS EC2 instances, plans and services.

## API

### GET /metadata

#### Response data

```
{
  User: {
    ID: string,
    Email: string
  },
  Services: [...],
  Servers: [...]
}
```

### POST /instance

#### Request body

```
{
  PlanID: uid
}
```

### GET /instance/:id

#### Response data

```
{
  PublicIp: string,
  LaunchTime: Date,
  State: string
}
```

## Models

### User

- ID: uid
- Email: string

### Plan

- ID: uid
- InstanceType: string
- Name: string
- VCores: number
- Memory: number
- Price: number
- Disk: number
- Enabled: boolean

### Services

- ID: string
- MinMem: number
- RecMem: number
- Name: string
- NameLong: string

### Server

- ID: uid
- ServiceID: string
- PlanID: uid
- UserID: uid

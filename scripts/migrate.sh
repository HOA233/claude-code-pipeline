#!/bin/bash
# Database Migration Script

set -e

MIGRATIONS_DIR="./migrations"
DRIVER="${DB_DRIVER:-redis}"
DB_URL="${DB_URL:-localhost:6379}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

usage() {
    echo "Usage: $0 [command]"
    echo ""
    echo "Commands:"
    echo "  create <name>   Create a new migration file"
    echo "  up              Apply all pending migrations"
    echo "  down            Rollback last migration"
    echo "  status          Show migration status"
    echo "  reset           Reset all migrations"
    echo ""
    echo "Environment:"
    echo "  DB_DRIVER   Database driver (default: redis)"
    echo "  DB_URL      Database URL (default: localhost:6379)"
}

create_migration() {
    local name=$1
    if [ -z "$name" ]; then
        echo "Error: Migration name required"
        usage
        exit 1
    fi

    local timestamp=$(date +%Y%m%d%H%M%S)
    local filename="${timestamp}_${name}.sh"

    mkdir -p "$MIGRATIONS_DIR"

    cat > "$MIGRATIONS_DIR/$filename" <<EOF
#!/bin/bash
# Migration: $name
# Created: $(date)

# Up migration
up() {
    echo "Applying migration: $name"
    # Add migration logic here
}

# Down migration
down() {
    echo "Rolling back migration: $name"
    # Add rollback logic here
}

# Run the appropriate function based on argument
case "\$1" in
    up) up ;;
    down) down ;;
    *) echo "Usage: \$0 [up|down]" ;;
esac
EOF

    chmod +x "$MIGRATIONS_DIR/$filename"
    echo -e "${GREEN}Created migration: $filename${NC}"
}

run_up() {
    echo -e "${YELLOW}Running migrations...${NC}"

    # Get applied migrations
    APPLIED=$(get_applied_migrations)

    for migration in $(ls "$MIGRATIONS_DIR"/*.sh 2>/dev/null | sort); do
        local name=$(basename "$migration")

        if echo "$APPLIED" | grep -q "$name"; then
            echo -e "  ${GREEN}✓${NC} $name (already applied)"
        else
            echo -e "  ${YELLOW}→${NC} Applying $name..."
            bash "$migration" up

            # Record migration
            record_migration "$name"

            echo -e "  ${GREEN}✓${NC} $name applied"
        fi
    done

    echo -e "${GREEN}Migrations complete${NC}"
}

run_down() {
    echo -e "${YELLOW}Rolling back last migration...${NC}"

    # Get last applied migration
    LAST=$(get_last_migration)

    if [ -z "$LAST" ]; then
        echo -e "${YELLOW}No migrations to rollback${NC}"
        return
    fi

    local migration="$MIGRATIONS_DIR/$LAST"

    if [ -f "$migration" ]; then
        echo -e "  ${YELLOW}→${NC} Rolling back $LAST..."
        bash "$migration" down

        # Remove record
        remove_migration "$LAST"

        echo -e "  ${GREEN}✓${NC} $LAST rolled back"
    fi

    echo -e "${GREEN}Rollback complete${NC}"
}

get_applied_migrations() {
    case $DRIVER in
        redis)
            redis-cli -u "$DB_URL" lrange migrations 0 -1 2>/dev/null || echo ""
            ;;
        *)
            echo ""
            ;;
    esac
}

get_last_migration() {
    case $DRIVER in
        redis)
            redis-cli -u "$DB_URL" lindex migrations -1 2>/dev/null || echo ""
            ;;
        *)
            echo ""
            ;;
    esac
}

record_migration() {
    local name=$1
    case $DRIVER in
        redis)
            redis-cli -u "$DB_URL" rpush migrations "$name" 2>/dev/null
            ;;
    esac
}

remove_migration() {
    local name=$1
    case $DRIVER in
        redis)
            redis-cli -u "$DB_URL" lrem migrations -1 "$name" 2>/dev/null
            ;;
    esac
}

show_status() {
    echo -e "${YELLOW}Migration Status:${NC}"
    echo ""

    APPLIED=$(get_applied_migrations)

    echo "Applied migrations:"
    if [ -z "$APPLIED" ]; then
        echo "  (none)"
    else
        echo "$APPLIED" | while read -r migration; do
            echo -e "  ${GREEN}✓${NC} $migration"
        done
    fi

    echo ""
    echo "Pending migrations:"
    local found=0

    for migration in $(ls "$MIGRATIONS_DIR"/*.sh 2>/dev/null | sort); do
        local name=$(basename "$migration")
        if ! echo "$APPLIED" | grep -q "$name"; then
            echo -e "  ${YELLOW}○${NC} $name"
            found=1
        fi
    done

    if [ $found -eq 0 ]; then
        echo "  (none)"
    fi
}

reset_migrations() {
    echo -e "${RED}Warning: This will reset all migrations!${NC}"
    read -p "Continue? (y/N) " confirm

    if [ "$confirm" != "y" ]; then
        echo "Cancelled"
        exit 0
    fi

    # Get all applied migrations
    APPLIED=$(get_applied_migrations)

    # Rollback in reverse order
    for migration in $(echo "$APPLIED" | tac); do
        local file="$MIGRATIONS_DIR/$migration"
        if [ -f "$file" ]; then
            echo "Rolling back $migration..."
            bash "$file" down
        fi
    done

    # Clear migration records
    case $DRIVER in
        redis)
            redis-cli -u "$DB_URL" del migrations 2>/dev/null
            ;;
    esac

    echo -e "${GREEN}All migrations reset${NC}"
}

# Main
case "${1:-}" in
    create)
        create_migration "$2"
        ;;
    up)
        run_up
        ;;
    down)
        run_down
        ;;
    status)
        show_status
        ;;
    reset)
        reset_migrations
        ;;
    *)
        usage
        ;;
esac
# dtui — Docker/Podman TUI Manager
# Recipes split by function into scripts/*.just modules.
# Run `just` or `just list` to see all available recipes.

mod build "scripts/build.just"
mod test "scripts/test.just"
mod verify "scripts/verify.just"
mod run "scripts/run.just"
mod dev "scripts/dev.just"
mod musl-gpgme "scripts/musl-gpgme/justfile"

# Show all recipes grouped by module
list:
    @just --list --unsorted

# Default recipe
default: build::build

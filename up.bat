@echo off

bun update --latest

cd native
cargo upgrade --incompatible
cargo update
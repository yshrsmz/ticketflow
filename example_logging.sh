#!/bin/bash

echo "Example: Using ticketflow with structured logging"
echo "================================================="
echo

echo "1. Running with default (silent) mode:"
./ticketflow list --status todo

echo
echo "2. Running with info-level logging to stderr:"
./ticketflow list --status todo --log-level info --log-format text

echo
echo "3. Running with debug-level JSON logging:"
./ticketflow list --status todo --log-level debug --log-format json

echo
echo "4. Running with logging to a file:"
./ticketflow list --status todo --log-level info --log-output ticketflow.log
echo "Check ticketflow.log for the output"
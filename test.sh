#!/bin/bash
# Test script for minishell

echo "Testing minishell..."
echo ""

# Test basic commands
echo "=== Test 1: Basic builtin commands ==="
echo "pwd"
echo "echo Hello World"
echo ""

# Test environment variables
echo "=== Test 2: Environment variables ==="
echo "echo \$HOME"
echo "echo Current user: \$USER"
echo ""

# Test pipelines
echo "=== Test 3: Pipelines ==="
echo "echo 'hello world' | grep hello"
echo "ps aux | head -5"
echo ""

# Test redirections
echo "=== Test 4: Redirections ==="
echo "echo 'test content' > test_output.txt"
echo "cat < test_output.txt"
echo ""

# Test conditional execution
echo "=== Test 5: Conditional execution ==="
echo "echo 'first' && echo 'second'"
echo "false || echo 'this should print'"
echo ""

# Test cd
echo "=== Test 6: cd command ==="
echo "cd /tmp && pwd"
echo "cd ~ && pwd"
echo ""

echo "exit"


#!/bin/bash

# Kiểm tra tham số message commit
if [ -z "$1" ]; then
    echo "Vui lòng nhập message commit."
    exit 1
fi

# Thêm tất cả thay đổi
git add .

# Commit có ký GPG với message truyền vào
git commit -S -m "$1"

# Đẩy lên nhánh hiện tại (main hoặc master)
branch=$(git rev-parse --abbrev-ref HEAD)
git push origin "$branch"


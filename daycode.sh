#!/bin/bash

# Thiết lập biến môi trường để tránh lỗi ký GPG
export GPG_TTY=$(tty)

# Kiểm tra tham số message commit
if [ -z "$1" ]; then
    echo "Vui lòng nhập message commit."
    exit 1
fi

# Thêm tất cả thay đổi
git add .

# Commit có ký GPG với message truyền vào
git commit -S -m "$1"

# Lấy tên nhánh hiện tại
branch=$(git rev-parse --abbrev-ref HEAD)

# Đẩy lên nhánh hiện tại
git push origin "$branch"

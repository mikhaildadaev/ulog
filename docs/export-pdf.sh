#!/bin/bash

# Наименование пакета
TITLE='ulog'

# Цвета для вывода
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Извлекаем версию из go.md
get_version() {
    local lang=$1
    local go_file="./${lang}/go.md"
    if [ -f "$go_file" ]; then
        grep -oE 'v[0-9]+\.[0-9]+\.[0-9]+' "$go_file" | head -1
    fi
}

# Порядок файлов для каждого языка
FILES=(
    "license.md"
    "go.md"
    "core_constructors.md"
    "core_options.md"
    "core_types.md"
    "sinkfile_constructors.md"
    "sinkfile_params.md"
    "sinkhttp_constructors.md"
    "sinkhttp_factories.md"
    "sinkhttp_params.md"
)

# Функция сборки PDF для одного языка
build_pdf() {
    local lang=$1
    local output=$2
    shift 2
    local files=("$@")

    local combined="combined-${lang}.md"

    echo -e "${YELLOW}Сборка PDF для языка: ${lang}${NC}"

    # Очищаем или создаём временный файл
    > "$combined"

    # Добавляем каждый файл в нужном порядке
    for file in "${files[@]}"; do
        local full_path="./${lang}/${file}"
        if [ -f "$full_path" ]; then
            echo "  ✓ Добавлен: ${file}"
            sed -e '/^---$/,/^---$/d' -e '/^:::/d' "$full_path" >> "$combined"
            # Добавляем разрыв страницы между файлами
            printf "\n\n<div style=\"page-break-before: always;\"></div>\n\n" >> "$combined"
        else
            echo -e "  ${RED}✗ Файл не найден: ${file}${NC}"
        fi
    done

    # Конвертируем в PDF
    echo "  Конвертация в PDF..."
    if npx md2pdf-context "$combined" -o "$output"; then
        echo -e "${GREEN}  ✓ Готово: ${output}${NC}"
    else
        echo -e "${RED}  ✗ Ошибка конвертации${NC}"
    fi

    # Удаляем временный файл
    rm -f "$combined"
    echo ""
}

# Собираем PDF для каждого языка
for lang in en ru zh; do
    version=$(get_version "$lang")
    if [ -z "$version" ]; then
        version="dev"
    fi
    output="public/${TITLE}-${version}-${lang}.pdf"
    echo -e "${YELLOW}Версия для ${lang}: ${version}${NC}"
    build_pdf "$lang" "$output" "${FILES[@]}"
done

echo -e "${GREEN}Все PDF собраны!${NC}"
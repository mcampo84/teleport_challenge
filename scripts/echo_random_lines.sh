#!/bin/bash

# List of 50 English words
WORDS=("apple" "banana" "cherry" "date" "elderberry" "fig" "grape" "honeydew" "kiwi" "lemon"
       "mango" "nectarine" "orange" "papaya" "quince" "raspberry" "strawberry" "tangerine" "ugli" "vanilla"
       "watermelon" "xigua" "yellow" "zucchini" "avocado" "blueberry" "cantaloupe" "dragonfruit" "eggplant" "fennel"
       "guava" "huckleberry" "jackfruit" "kumquat" "lime" "mulberry" "olive" "peach" "plum" "pomegranate"
       "rhubarb" "starfruit" "tomato" "yam" "zebra" "apricot" "blackberry" "coconut" "durian" "elderflower")

# Function to generate a random line of text with a specified number of words
generate_random_line() {
  local num_words=$((RANDOM % 9 + 7)) # Generate a random number between 7 and 15
  local line=""
  for ((i = 0; i < num_words; i++)); do
    line+="${WORDS[RANDOM % ${#WORDS[@]}]} "
  done
  echo "$line"
}

# Echo random lines of text at a rate of one line per second for five minutes
for ((i = 0; i < 300; i++)); do
  generate_random_line
  sleep 1
done

#!/usr/bin/env python3
from os.path import exists
from googlerobotstxt import allowed_by_robots

def insert_into_csv_at(index: int, insert: str, into: str):
    """Inserts a single value into a given .csv formatted line."""
    lwords = into.split(',')
    lwords.insert(index, insert) # Write column names to combined output
    modded_line = ','.join(map(str, lwords))
    return modded_line

# Constants
DEFAULT_USERAGENT_TOKENS = ["SiteimproveBot", "SiteimproveBot-Crawler"]
ID_IDX = 0
END_URL_IDX = 3
STATUS_IDX = 4

# Paths
run_dir = "../runs/4_run/"
input_path = run_dir + "success.csv"
output_path = run_dir + "success_robots.csv"
robots_dir = run_dir + "robots/"
parse_errors = []

input_file = open(input_path, "r")
output_file = open(output_path, 'a')

# Add entry to column names on first row
column_names = input_file.readline()
output_file.write(insert_into_csv_at(STATUS_IDX+1, "Robots.txt OK", column_names))

print("Checking robots.txt of URL's from {0}".format(input_path))
print("Saving to {0}".format(output_path))
while True:
    line = input_file.readline()
    if not line:
        break

    entries_suc = line.split(',') # See format in pkg/work/work.go
    id_suc = entries_suc[ID_IDX]
    end_url_suc = entries_suc[END_URL_IDX]

    # Iterate over robot files using succes id
    robots_txt_file = robots_dir + id_suc + '.rob'
    if exists(robots_txt_file):
        try:
            # Alternatively use ISO-8859-1 or Latin_1 is 8 bit character sets, so all garbage has a valid value. 
            with open(robots_txt_file, 'r', encoding='utf-8') as robotsfile:
                contents = robotsfile.read() # Can potentially fail with Decoding Error
                is_allowed = allowed_by_robots(contents, user_agents=DEFAULT_USERAGENT_TOKENS, url=end_url_suc)
                output_file.write(insert_into_csv_at(STATUS_IDX+1, str(is_allowed), line))
        except UnicodeDecodeError as e:
            error = "error: robots.txt with id {0} is not UTF-8 encoded: {1}".format(id_suc, e)
            output_file.write(insert_into_csv_at(STATUS_IDX+1, "Error", line))
            parse_errors.append(error)
            continue
    else:
        # Add line with empty robots.txt if not present
        output_file.write(insert_into_csv_at(STATUS_IDX+1, "None", line))


input_file.close()
output_file.close()

print("Parse errors encountered: {0}".format(len(parse_errors)))
for error in parse_errors:
    print(error)

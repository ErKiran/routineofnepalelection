"""
Script to scrape news from www.kantipurdaily.com/news
Contains:
    kantipur_daily_extractor(): Gives list of news dicts
"""
import re
import json
from bs4 import BeautifulSoup as BS
import requests
import sys


parser = "lxml"

regex = re.compile(
    r'^(?:http|ftp)s?://'  # http:// or https://
    # domain...
    r'(?:(?:[A-Z0-9](?:[A-Z0-9-]{0,61}[A-Z0-9])?\.)+(?:[A-Z]{2,6}\.?|[A-Z0-9-]{2,}\.?)|'
    r'localhost|'  # localhost...
    r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})'  # ...or ip
    r'(?::\d+)?'  # optional port
    r'(?:/?|[/?]\S+)$', re.IGNORECASE)


def setup(url):
    headers = {
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_1) AppleWebKit\
        /537.36(KHTML, like Gecko) Chrome/39.0.2171.95 Safari/537.36"
    }
    try:
        page = requests.get(url, headers=headers)
    except Exception as e:
        print("Connection refused by the server..", e)
    soup = BS(page.content, parser)
    return soup


def kantipur_election(url):
    soup = setup(url)
    counter = 0
    news_list = []
    from pprint import pprint
    title = soup.find("h2", {"class": "header-title"}).text
    for article in soup.find_all("div", class_="candidate-list"):
        name = article.find('div', class_='candidate-name').text
        party = article.find('div', class_='candidate-party-name').text
        vote_no = article.find('div', class_='vote-numbers').text
        image_url = article.find(
            'div', class_='candidate-img').find('img')['src']
        election_icon = article.find(
            'div', class_='election-icon').find('img')['src']

        t = (re.match(regex, election_icon) is not None)
        if t == False:
            election_icon = "https://election.ekantipur.com" + election_icon

        if image_url == "":
            image_url = "https://www.seekpng.com/png/full/966-9665317_placeholder-image-person-jpg.png"
        data_dict = {
            "candidateName": name,
            "candidatePartyName": party,
            "voteNumbers": vote_no,
            "candidateImage": image_url,
            "electionIcon": election_icon,
            "title": title
        }
        news_list.append(data_dict)
        counter += 1

    pprint(news_list[0:3])

    with open("data/"+sys.argv[1].strip()+".json", "w") as outfile:
        json.dump(news_list[0:3], outfile)
    return news_list[0:3]


if __name__ == "__main__":
    with open('data.json', 'r') as rf:
        data = json.load(rf)
        cities = data.keys()
    result = kantipur_election(data[sys.argv[1]])

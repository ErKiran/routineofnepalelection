import facebook as fb
import json
import sys

token = ""
def posttofacebook(city):
    fbobj = fb.GraphAPI(token)
    caption = ""
    hastags = ["#routineofnepalelection",
               "#voteforchange2079", "#voteforchange2022", "#liveelectionupdate", "#electionupdate", "#votecountlive"]

    # caption += result[0]['title'].replace(" ", "_") + "\n"
    caption += "Live Election Update Nepal\n"
    caption += "Election update of " + city + "\n"
    caption += "Local Election 2079 Update\n"
    caption += "Local Election 2022 Update\n"
    caption += "Vote Count Live\n"
    caption += "Election Update Nepal\n"
    caption += "Routine of Nepal Election\n"
    with open("data/"+city+".json", "r") as rf:
        data = json.load(rf)
        for i in data:
            caption += "#" + i['candidateName'].replace(" ", "_") + "\n" + \
                "#" + i['candidatePartyName'].replace(" ", "_") + "\n"
    for hashtag in hastags:
        caption += hashtag + "\n"

    fbobj.put_photo(image=open('city.png', 'rb'), message=caption)


if __name__ == "__main__":
    posttofacebook(sys.argv[1])

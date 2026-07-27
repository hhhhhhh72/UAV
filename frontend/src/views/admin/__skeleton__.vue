#!/bin/bash
SRC="__skeleton__.vue"
BASE="D:/w-yao/frontend/src/views/admin"
cd "$BASE"

declare -A MODS
MODS=(
  ["shops/ShopList.vue"]="shops"
  ["experts/ExpertList.vue"]="experts"
  ["resources/ResourceList.vue"]="resources"
  ["compliance/ComplianceList.vue"]="compliance"
  ["training/CourseList.vue"]="training"
  ["training/CertList.vue"]="certs"
  ["jobs/JobList.vue"]="jobs"
  ["colleges/CollegeList.vue"]="colleges"
  ["study/StudyList.vue"]="study"
  ["achievements/AchievementList.vue"]="achievements"
  ["challenges/ChallengeList.vue"]="challenges"
  ["projects/ProjectList.vue"]="projects"
  ["testsites/TestSiteList.vue"]="testsites"
  ["transformations/TransList.vue"]="transformations"
  ["events/EventList.vue"]="events"
  ["portfolios/PortfolioList.vue"]="portfolios"
  ["exhibitions/ExhibitionList.vue"]="exhibitions"
  ["reports/ReportList.vue"]="reports"
  ["emergency/ResourceList.vue"]="emergency-resources"
  ["emergency/DispatchList.vue"]="emergency-dispatches"
  ["messages/NotifyList.vue"]="messages"
)

for FILE in "${!MODS[@]}"; do
  API=${MODS[$FILE]}
  cp "$SRC" "$FILE"
  sed -i "s|PLACEHOLDER|$API|" "$FILE"
  echo "OK $FILE"
done

rm "$SRC"
echo "DONE: ${#MODS[@]} files"

http://localhost:5078/TestResults/sample/v10.pw.StaffNavigation.I_navigate_to_all_staff_menu_items_in_Continuum.zip

trx log
junit log
playwright trace

general image & pdf viewers
what about coverage testing formats?
what about generalized logging for associated server logs?

vitest has a json format

https://www.nuget.org/packages/JUnitXml.TestLogger

to output in JUnit format you can add a reference to the JUnit Logger nuget package in your test project and then use the following command to run tests
dotnet test --test-adapter-path:. --logger:junit

See more usage details: https://github.com/spekt/junit.testlogger#usage
 "Publish JUnit report" post-build action in Jenkins.

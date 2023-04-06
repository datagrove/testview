using Microsoft.Extensions.FileProviders;
using System.Text.Json;

internal class Program
{
    static Show3 show3  = new Show3("","");
    static string show3dir =  "" ;
    static string resolve(string e)
    {
       return Path.GetFullPath(Path.Join(show3dir, e));
    }
    private static void Main(string[] args)
    {
        show3dir = args.Count() > 0 ? args[0] : Environment.CurrentDirectory;
        var read = JsonSerializer.Deserialize<Show3>(File.ReadAllText(Path.Join(show3dir, "show3.json")));
        if (read == null)
        {
            Console.WriteLine("no show3.json");
            Environment.Exit(1);
        }
        show3 = new Show3(resolve(read.testRoot), resolve(read.appRoot));
        Console.WriteLine(show3);



        var builder = WebApplication.CreateBuilder(new WebApplicationOptions()
        {
            Args = args,
            EnvironmentName = "Development",
            WebRootPath = show3.appRoot,   
        });


        var app = builder.Build();
        var options = new DefaultFilesOptions();
        options.DefaultFileNames.Add("index.html"); // shouldn't need this?
        app.UseDefaultFiles(options); // call before UseStaticFiles
        app.UseStaticFiles(new StaticFileOptions
        {
            FileProvider = new PhysicalFileProvider(show3.appRoot),
            RequestPath = "",
            ServeUnknownFileTypes=true
        });
        app.UseStaticFiles(new StaticFileOptions
        {
            FileProvider = new PhysicalFileProvider(show3.testRoot),
            RequestPath = "/TestResults",
            ServeUnknownFileTypes=true
        });

        app.MapGet("/api/runs", async () =>
        {
            await Task.CompletedTask;
            List<string> dir = new();
            foreach (var batch in Directory.GetDirectories(show3.testRoot))
            {
                dir.Add(Path.GetFileName(batch));
            }

            return JsonSerializer.Serialize(dir, new JsonSerializerOptions { WriteIndented = true });

        });

        // get summary of one run, may be in progress
        // the core thing we serve is in index.json, which is created at the beginning of the test, but it must be merged with the status of files in the directory as status (pass/fail/waiting)
        app.MapGet("/api/run/{batch}", async (string batch) =>
        {

            await Task.CompletedTask;
            batch = Path.Join(show3.testRoot, batch);
            // index.json written at beginning of each test
            List<string>? root = JsonSerializer.Deserialize<List<string>>(File.ReadAllText(Path.Join(batch, "index.json")));
            if (root == null)
            {
                throw new Exception("no test definition");
            }
            Dictionary<string, string> testcode = new();
            foreach (var feature in root)
            {
                testcode[feature] = "waiting";
            }

            HashSet<string> failed = new();

            // I don't need this, web can just look for the file
            foreach (var f in Directory.GetFiles(batch))
            {
                var fn = Path.GetFileName(f);
                var p = fn.Split(".");
                var ext = p.Last();
                var tn = string.Join(".", p.Take(4));
                switch (ext)
                {
                    // case "zip":
                    //     var key = string.Join(".", p.Take(p.Count() - 1));
                    //     break;
                    case "txt": // error
                        if ("error" == p[p.Count() - 2])
                            failed.Add(tn);
                        else
                        {// test file\b
                            testcode[tn] = "pass";
                        }
                        break;
                }
            }
            foreach (var key in failed)
            {
                testcode[key] = "fail";
            }

            return JsonSerializer.Serialize(testcode);
        });



        app.Run("http://localhost:5078");
    }
}

public class Show3 {
    public string testRoot { get; set; }
    public string appRoot { get; set; }
    public Show3(string testRoot, string appRoot)
    {
        this.testRoot = testRoot;
        this.appRoot = appRoot;
    }
    public override string ToString()
    {
        return JsonSerializer.Serialize(this, new JsonSerializerOptions { WriteIndented = true });
    }
};
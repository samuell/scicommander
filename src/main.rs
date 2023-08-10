use clap::Parser;

#[derive(Parser)]
struct Cli {
    command: String,
}

fn main() {
    let args = Cli::parse();
    println!("Command: {}", args.command)
}

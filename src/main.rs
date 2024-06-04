use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::Deserialize;
use std::collections::{HashMap, HashSet, VecDeque};

#[derive(Deserialize)]
struct Pair {
    from: String,
    to: String,
}

#[derive(Deserialize)]
struct Pairs {
    pairs: Vec<Pair>,
}

#[post("/api/submit_pairs")]
async fn submit_pairs(pairs: web::Json<Pairs>) -> impl Responder {
    let pairs = &pairs.pairs;
    let mut graph: HashMap<String, Vec<String>> = HashMap::new();
    let mut in_degree: HashMap<String, usize> = HashMap::new();

    for pair in pairs {
        graph.entry(pair.from.clone()).or_default().push(pair.to.clone());
        *in_degree.entry(pair.to.clone()).or_default() += 1;
        in_degree.entry(pair.from.clone()).or_default();
    }

    match topological_sort(&graph, &in_degree) {
        Ok(sorted) => HttpResponse::Ok().json(sorted),
        Err((most_outgoing, least_incoming)) => HttpResponse::Ok().json(format!(
            "Cycle detected. Break cycle by removing edge from '{}' (most outgoing) to '{}' (least incoming).",
            most_outgoing, least_incoming
        )),
    }
}

fn topological_sort(
    graph: &HashMap<String, Vec<String>>,
    in_degree: &HashMap<String, usize>,
) -> Result<Vec<String>, (String, String)> {
    let mut queue: VecDeque<String> = VecDeque::new();
    let mut in_degree = in_degree.clone();
    let mut sorted: Vec<String> = Vec::new();
    let mut max_outgoing = ("".to_string(), 0);
    let mut min_incoming = ("".to_string(), usize::MAX);

    for (node, &degree) in &in_degree {
        if degree == 0 {
            queue.push_back(node.clone());
        }
        if graph.get(node).unwrap_or(&vec![]).len() > max_outgoing.1 {
            max_outgoing = (node.clone(), graph.get(node).unwrap_or(&vec![]).len());
        }
        if degree < min_incoming.1 {
            min_incoming = (node.clone(), degree);
        }
    }

    while let Some(node) = queue.pop_front() {
        sorted.push(node.clone());
        if let Some(neighbors) = graph.get(&node) {
            for neighbor in neighbors {
                let entry = in_degree.entry(neighbor.clone()).or_default();
                *entry -= 1;
                if *entry == 0 {
                    queue.push_back(neighbor.clone());
                }
            }
        }
    }

    if sorted.len() == graph.len() {
        Ok(sorted)
    } else {
        Err((max_outgoing.0, min_incoming.0))
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    HttpServer::new(|| {
        App::new()
            .service(submit_pairs)
    })
    .bind("127.0.0.1:8080")?
    .run()
    .await
}

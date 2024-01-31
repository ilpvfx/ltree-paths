use std::str::FromStr;

use criterion::{criterion_group, criterion_main, BenchmarkId, Criterion};

use ltree_rust::{paths::LTreePathBuf, LTreePath};
use rand::{
    distributions::{Alphanumeric, DistString},
    rngs::SmallRng,
    SeedableRng,
};

fn generate_string_label(rng: &mut SmallRng) -> String {
    const MAX_LEN: usize = 16;
    Alphanumeric.sample_string(rng, MAX_LEN)
}

fn generate_string_path(rng: &mut SmallRng, n: i32) -> String {
    let mut labels: Vec<String> = Vec::default();

    for _ in 0..n {
        labels.push(generate_string_label(rng));
    }

    labels.join(LTreePath::DELIMITER_STR)
}

fn parsing(c: &mut Criterion) {
    let mut g = c.benchmark_group("parsing");
    let mut rng = SmallRng::seed_from_u64(12345);

    for n in 1..9 {
        g.bench_with_input(BenchmarkId::from_parameter(n), &n, |b, n| {
            let generator = || generate_string_path(&mut rng, *n);
            let test = |str_path| LTreePath::try_from(str_path).expect("a valid path");

            b.iter_batched(generator, test, criterion::BatchSize::SmallInput)
        });
    }
}

fn joining(c: &mut Criterion) {
    let mut g = c.benchmark_group("joining");
    let mut rng = SmallRng::seed_from_u64(12345);

    let typical_expected_path =
        LTreePath::from_str("widget-company.products.year.2024.ᛒluetoothSpeaker")
            .expect("a valid path");

    for n in 1..9 {
        g.bench_with_input(
            BenchmarkId::new("medium_length_path_join_path", n),
            &n,
            |b, n| {
                let generator = || {
                    LTreePath::try_from(generate_string_path(&mut rng, *n)).expect("a valid path")
                };
                let test = |str_path| typical_expected_path.join(&str_path).expect("a valid path");

                b.iter_batched(generator, test, criterion::BatchSize::SmallInput)
            },
        );
    }

    for n in 1..9 {
        g.bench_with_input(BenchmarkId::new("n_labels_join", n), &n, |b, n| {
            let generator = || {
                let mut labels: Vec<LTreePath> = Vec::new();
                for _ in 0..*n {
                    labels.push(generate_string_label(&mut rng).try_into().unwrap());
                }

                labels
            };
            let test = |labels: Vec<LTreePath>| {
                labels
                    .into_iter()
                    .reduce(|a, b| a.join(&b).unwrap())
            };

            b.iter_batched(generator, test, criterion::BatchSize::SmallInput)
        });
    }
}

fn buf(c: &mut Criterion) {
    let mut g = c.benchmark_group("buf");
    let mut rng = SmallRng::seed_from_u64(12345);

    for n in 1..9 {
        g.bench_with_input(BenchmarkId::new("n_labels_to_ltree_path", n), &n, |b, n| {
            let generator = || {
                let mut buf = LTreePathBuf::new();
                for _ in 0..*n {
                    buf.push_label(generate_string_label(&mut rng))
                        .unwrap();
                }

                buf
            };
            let test = |buf: LTreePathBuf| buf.to_ltree_path();

            b.iter_batched(generator, test, criterion::BatchSize::SmallInput)
        });
    }
}

criterion_group!(benches, parsing, joining, buf);
criterion_main!(benches);

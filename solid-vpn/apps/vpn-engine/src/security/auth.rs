use tracing::warn;

pub fn validate_token(provided: &str, expected: &str) -> bool {
    use std::hint::black_box;
    if provided.len() != expected.len() {
        warn!("token length mismatch");
        return false;
    }
    let result = provided
        .bytes()
        .zip(expected.bytes())
        .fold(0u8, |acc, (a, b)| acc | (a ^ b));
    black_box(result) == 0
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn valid_token_accepted() {
        assert!(validate_token("correct-token", "correct-token"));
    }

    #[test]
    fn invalid_token_rejected() {
        assert!(!validate_token("wrong-token--", "correct-token"));
    }

    #[test]
    fn different_length_rejected() {
        assert!(!validate_token("short", "a-much-longer-token"));
    }
}

import 'package:uuid/uuid.dart';

const success = 'success';
const error = 'error';
const lastRefreshedAt = 'last_refreshed_at';
const accessToken = 'access_token';
const refreshToken = 'refresh_token';
const permissions = 'permissions';

const nodeCode = '''
import React from "react";

// this is a comment

export default function TestimonialItem({ testimonial }: TestimonialProps) {
    return (
        <div className="group">
            <p className="mb-4 text-lg font-medium group-hover:text-white">
                {testimonial.testimonial}
            </p>
            <p className="font-semibold text-slate-600 group-hover:text-white">
                {`— {testimonial.name}`}
            </p>
            <p className="font-semibold text-slate-600 group-hover:text-white">
                {`— {testimonial.name}`}
            </p>
        </div>
    );
}''';

const nodeResult = r'''
Running with gitlab-runner 12.3.0-rc1
$ sh build.sh This is output from build.sh Job succeeded
''';

const nodeWidth = 240.0;
const nodeHeight = 160.0;

String generateId() => const Uuid().v1();
